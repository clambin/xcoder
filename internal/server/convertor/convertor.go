package convertor

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/clambin/vidconv/internal/server/requests"
	"github.com/clambin/vidconv/pkg/ffmpeg"
	"github.com/go-chi/chi/v5"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

type VideoConvertor interface {
	Convert(ctx context.Context, request ffmpeg.Request) error
}

type Convertor struct {
	VideoConvertor
	RemoveConverted bool
	requests        *requests.Requests
	logger          *slog.Logger
	active          atomic.Bool
	stats           stats
}

func New(req *requests.Requests, removeConverted bool, logger *slog.Logger) *Convertor {
	return &Convertor{
		VideoConvertor:  ffmpeg.Processor{Logger: logger.With("component", "ffmpeg")},
		RemoveConverted: removeConverted,
		requests:        req,
		logger:          logger,
	}
}

const refreshInterval = 500 * time.Millisecond
const reportInterval = time.Minute

func (c *Convertor) Run(ctx context.Context, concurrent int) error {
	c.logger.Debug("started")
	defer c.logger.Debug("stopped")

	var wg sync.WaitGroup
	wg.Add(concurrent)

	for i := 0; i < concurrent; i++ {
		go func() {
			defer wg.Done()
			_ = c.run(ctx)
		}()
	}

	wg.Wait()
	return nil
}

func (c *Convertor) run(ctx context.Context) error {
	c.logger.Debug("started")
	defer c.logger.Debug("stopped")

	for {
		if c.active.Load() {
			if request, ok := c.requests.GetNext(); ok {
				if err := c.process(ctx, request); err != nil {
					c.logger.Error("conversion failed", slog.Any("err", err))
					c.requests.Add(request)
				}
			}
		}
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(refreshInterval):
		}
	}
}

func (c *Convertor) process(ctx context.Context, request requests.Request) error {
	c.stats.push(request.Source)
	defer c.stats.pop(request.Source)

	start := time.Now()
	lastReport := start

	progressCallback := func(p ffmpeg.Progress) {
		if time.Since(lastReport) > reportInterval {
			c.logger.Info("converting video",
				slog.String("input", request.Source),
				slog.String("converted", fmt.Sprintf("%.1f%%", 100*float64(p.Converted)/float64(request.SourceStats.Duration()))),
				slog.Float64("speed", p.Speed),
				slog.Duration("remaining", calculateRemaining(start, time.Now(), p.Converted, request.SourceStats.Duration())),
			)
			lastReport = time.Now()
		}
	}

	c.logger.Info("converting video", slog.Any("request", request))

	err := c.VideoConvertor.Convert(
		ctx,
		ffmpeg.Request{
			Source:        request.Source,
			Target:        request.Target,
			VideoCodec:    request.VideoCodec,
			BitsPerSample: request.BitsPerSample,
			BitRate:       request.BitRate,
			ProgressCB:    progressCallback,
		},
	)
	if err != nil {
		if err := os.Remove(request.Target); err != nil {
			c.logger.Error("failed to clean up partially converted file", slog.Any("err", err))
		}
		return err
	}

	if compression, err := CalculateCompression(request.Source, request.Target); err == nil {
		c.logger.Info("video converted", slog.Any("request", request), slog.Any("compression", compression))
	} else {
		c.logger.Warn("video converted. failed to calculate compression factor", slog.Any("request", request))
	}

	if c.RemoveConverted {
		if err = os.Remove(request.Source); err != nil {
			return fmt.Errorf("remove original file: %w", err)
		}
	}

	return nil
}

func (c *Convertor) Router(r chi.Router) {
	r.Route("/active", func(r chi.Router) {
		r.Get("/", c.HandlerIsActive)
		r.Post("/on", c.HandlerActiveOn)
		r.Post("/off", c.HandlerActiveOff)
	})
}

func (c *Convertor) SetActive(active bool) {
	c.active.Store(active)
}

func (c *Convertor) HandlerIsActive(w http.ResponseWriter, _ *http.Request) {
	body := struct {
		Active bool
	}{
		Active: c.active.Load(),
	}

	_ = json.NewEncoder(w).Encode(body)
}

func (c *Convertor) HandlerActiveOn(_ http.ResponseWriter, _ *http.Request) {
	c.SetActive(true)
}

func (c *Convertor) HandlerActiveOff(_ http.ResponseWriter, _ *http.Request) {
	c.SetActive(false)
}

func (c *Convertor) Health(_ context.Context) any {
	processing := c.stats.getStats()

	return struct {
		Active     bool
		Processing []string
	}{
		Active:     c.active.Load(),
		Processing: processing,
	}
}
