package vidconv

import (
	"context"
	"fmt"
	"github.com/clambin/vidconv/internal/feeder"
	"github.com/clambin/vidconv/internal/health"
	"github.com/clambin/vidconv/internal/inspector"
	"github.com/clambin/vidconv/internal/processor"
	"log/slog"
	"sync"
	"time"
)

type Application struct {
	Feeder    *feeder.Feeder
	Inspector *inspector.Inspector
	Processor *processor.Processor
	addr      string
	logger    *slog.Logger
}

const scanInterval = time.Hour

func New(rootDir string, targetCodec string, addr string, logger *slog.Logger) *Application {
	f := feeder.New(rootDir, scanInterval, logger.With(slog.String("component", "feeder")))
	s := inspector.New(f.Feed, targetCodec, logger.With(slog.String("component", "inspector")))
	p := processor.New(s.Output, targetCodec, logger.With(slog.String("component", "processor")))
	return &Application{
		Feeder:    f,
		Inspector: s,
		Processor: p,
		addr:      addr,
		logger:    logger,
	}
}

func (a Application) Run(ctx context.Context, concurrent int) error {
	h := health.Health{
		Addr:       ":9090",
		Components: []health.HealthChecker{a.Processor},
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = h.Run(ctx)
	}()

	a.logger.Info("starting convertors", "count", concurrent)
	wg.Add(concurrent)
	for i := 0; i < concurrent; i++ {
		go func() {
			defer wg.Done()
			if err := a.Processor.Run(ctx); err != nil {
				a.logger.Error("convertor failed", "err", err)
			}
		}()
	}
	wg.Add(concurrent)
	for j := 0; j < concurrent; j++ {
		go func() {
			defer wg.Done()
			if err := a.Inspector.Run(ctx); err != nil {
				a.logger.Error("inspector failed", "err", err)
			}
		}()
	}
	if err := a.Feeder.Run(ctx); err != nil {
		return fmt.Errorf("feeder failed: %w", err)
	}

	// allow processors to process their last file
	for {
		time.Sleep(time.Second)
		if h := a.Processor.Health(ctx).(processor.Health); len(h.Processing) == 0 {
			break
		}
	}

	return nil
}
