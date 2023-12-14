package scanner

import (
	"context"
	"fmt"
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

const scanInterval = time.Hour

func New(rootDir string, profile string, r *requests.Requests, logger *slog.Logger) *Scanner {
	f := feeder.New(rootDir, scanInterval, logger.With(slog.String("component", "feeder")))
	return &Scanner{
		Feeder:    f,
		Inspector: inspector.New(f.Feed, profile, r, logger.With(slog.String("component", "inspector"))),
		requests:  r,
		logger:    logger,
	}
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
	start := time.Now()
	defer func(start time.Time) {
		a.logger.Info("scan done", slog.Duration("duration", time.Since(start)))
	}(start)
	if err := a.Feeder.Run(ctx); err != nil {
		return fmt.Errorf("feeder failed: %w", err)
	}
	return nil
}

func (a Scanner) Health(_ context.Context) any {
	return struct {
		Queued []string
	}{
		Queued: a.requests.List(),
	}
}
