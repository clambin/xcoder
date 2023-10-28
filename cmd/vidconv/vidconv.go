package main

import (
	"context"
	"flag"
	"github.com/clambin/vidconv/internal/feeder"
	"github.com/clambin/vidconv/internal/health"
	"github.com/clambin/vidconv/internal/processor"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var (
	debug      = flag.Bool("debug", false, "switch on debug logging")
	input      = flag.String("input", "/Volumes/Video/movies", "input directory")
	convertors = flag.Int("convertors", 1, "number of video convertors ")
)

func main() {
	flag.Parse()

	var handlerOpts slog.HandlerOptions
	if *debug {
		handlerOpts.Level = slog.LevelDebug
	}

	l := slog.New(slog.NewTextHandler(os.Stderr, &handlerOpts))
	s := feeder.New(*input, time.Second, l.With("component", "scanner"))
	p := processor.New(s.Feed, l.With("component", "processor"), "hevc")

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	h := health.Health{
		Addr:       ":9090",
		Components: []health.HealthChecker{p},
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = h.Run(ctx)
	}()

	l.Info("starting convertors", "count", *convertors)
	wg.Add(*convertors)
	for i := 0; i < *convertors; i++ {
		go func() {
			defer wg.Done()
			if err := p.Run(ctx); err != nil {
				l.Error("convertor failed", "err", err)
			}
		}()
	}
	if err := s.Run(ctx); err != nil {
		l.Error("feeder failed", "err", err)
	}

	wg.Wait()
}
