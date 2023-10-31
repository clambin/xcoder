package processor

import (
	"context"
	"errors"
	"github.com/clambin/vidconv/internal/health"
	"github.com/clambin/vidconv/internal/inspector"
	"github.com/clambin/vidconv/pkg/ffmpeg"
	"github.com/clambin/vidconv/pkg/syncset"
	"log/slog"
	"os"
	"sync/atomic"
	"time"
)

// A Processor converts a video
type Processor struct {
	VideoConvertor
	Source     <-chan inspector.ConversionRequest
	Logger     *slog.Logger
	received   atomic.Int64
	accepted   atomic.Int64
	processing *syncset.Set[string]
}

var _ health.HealthChecker = &Processor{}

type VideoConvertor interface {
	ConvertWithProgress(ctx context.Context, input string, output string, targetCodec string, bitrate int, progressCallback func(probe ffmpeg.Progress)) error
}

const (
	loggingInterval = 0.05
)

func New(s <-chan inspector.ConversionRequest, l *slog.Logger) *Processor {
	return &Processor{
		VideoConvertor: ffmpeg.Processor{Logger: l.With(slog.String("component", "ffmpeg"))},
		Source:         s,
		Logger:         l,
		processing:     syncset.New[string](),
	}
}

func (p *Processor) Run(ctx context.Context) error {
	p.Logger.Info("started")
	defer p.Logger.Info("stopped")
	for {
		select {
		case <-ctx.Done():
			return nil
		case request, ok := <-p.Source:
			if !ok {
				return nil
			}
			if err := p.process(ctx, request); err != nil {
				p.Logger.Error("conversion failed", "err", err, "path", request.Input.Path)
			}
		}
	}
}

func (p *Processor) process(ctx context.Context, request inspector.ConversionRequest) error {
	l := p.Logger.With(slog.String("path", request.Input.Path))
	l.Debug("file received")

	p.received.Add(1)

	l.Info("converting file",
		slog.Any("in", request.Input.Stats),
		slog.Group("out",
			slog.String("codec", request.TargetCodec),
			slog.Int("bitrate", request.TargetBitrate/1024),
		),
	)

	p.accepted.Add(1)
	p.processing.Add(request.Input.Path)
	defer p.processing.Remove(request.Input.Path)

	start := time.Now()

	if err := p.convert(ctx, l, request); err != nil {
		return err
	}

	compression, err := calculateCompression(request.Input.Path, request.TargetFile)
	if err != nil {
		l.Warn("failed to calculate compression ratio", "err", err)
	}

	if err = os.Remove(request.Input.Path); err != nil {
		l.Warn("failed to remove source file", "err", err)
	}

	l.Info("converted file",
		slog.Any("compression", compression),
		slog.Duration("duration", time.Since(start)),
	)
	return nil
}

func (p *Processor) convert(ctx context.Context, l *slog.Logger, request inspector.ConversionRequest) error {
	var lastRatio float64
	//var called, done int
	duration := request.Input.Stats.Duration()
	err := p.VideoConvertor.ConvertWithProgress(
		ctx, request.Input.Path,
		request.TargetFile, request.TargetCodec, request.TargetBitrate,
		func(progress ffmpeg.Progress) {
			//called++
			ratio := float64(progress.Converted) / float64(duration)
			if ratio-lastRatio < loggingInterval {
				return
			}
			//done++
			lastRatio = ratio
			remaining := duration - progress.Converted
			remaining = time.Duration(float64(remaining) / progress.Speed)
			l.Info("converting",
				slog.Int("progress(%)", int(100*ratio)),
				slog.Float64("speed", progress.Speed),
				slog.Duration("expected", remaining),
			)
		})

	if err != nil {
		if err2 := os.Remove(request.TargetFile); !errors.Is(err2, os.ErrNotExist) {
			l.Error("failed to remove partial output file", "err", err2)
		}
		return err
	}

	if err = os.Chown(request.TargetFile, 568, 568); err != nil {
		l.Warn("failed to set ownership", "err", err)
	}
	return nil
}
