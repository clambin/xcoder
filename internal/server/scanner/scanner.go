package scanner

import (
	"context"
	"github.com/clambin/videoConvertor/internal/server/requests"
	"github.com/clambin/videoConvertor/internal/server/scanner/feeder"
	"github.com/clambin/videoConvertor/internal/server/scanner/inspector"
	"log/slog"
	"sync"
	"time"
)

type Scanner struct {
	Feeder    *feeder.Feeder
	Inspector *inspector.Inspector
	requests  *requests.Requests
	logger    *slog.Logger
}

type Config struct {
	RootDir string
	Profile string
}

const scanInterval = time.Hour

func New(cfg Config, r *requests.Requests, logger *slog.Logger) (*Scanner, error) {
	f := feeder.New(cfg.RootDir, scanInterval, logger.With(slog.String("component", "feeder")))
	i, err := inspector.New(f.Feed, cfg.Profile, r, logger.With(slog.String("component", "inspector")))
	if err != nil {
		return nil, err
	}
	s := Scanner{
		Feeder:    f,
		Inspector: i,
		requests:  r,
		logger:    logger,
	}
	return &s, nil
}

func (a Scanner) Run(ctx context.Context, concurrent int) error {
	var wg sync.WaitGroup
	wg.Add(concurrent)
	for j := 0; j < concurrent; j++ {
		go func() {
			defer wg.Done()
			if err := a.Inspector.Run(ctx); err != nil {
				a.logger.Error("inspector failed", slog.Any("err", err))
			}
		}()
	}

	for {
		if err := a.scan(ctx); err != nil {
			a.logger.Error("scan failed", slog.Any("err", err))
		}
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(time.Hour):
		}
	}
}

func (a Scanner) scan(ctx context.Context) error {
	start := time.Now()
	defer func(start time.Time) {
		a.logger.Info("scan done", slog.Duration("duration", time.Since(start)))
	}(start)
	return a.Feeder.Run(ctx)
}

func (a Scanner) Health(_ context.Context) any {
	return struct {
		Queued []string
	}{
		Queued: a.requests.List(),
	}
}
