package main

import (
	"context"
	"flag"
	"github.com/clambin/vidconv/internal/cmd/vidconv"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

var (
	debug      = flag.Bool("debug", false, "switch on debug logging")
	input      = flag.String("input", "/Volumes/Video/movies", "input directory")
	concurrent = flag.Int("concurrent", 1, "number of concurrent video convertors ")
)

func main() {
	flag.Parse()

	var handlerOpts slog.HandlerOptions
	if *debug {
		handlerOpts.Level = slog.LevelDebug
	}

	l := slog.New(slog.NewTextHandler(os.Stderr, &handlerOpts))
	a := vidconv.New(*input, "hevc", ":9090", l)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err := a.Run(ctx, *concurrent); err != nil {
		l.Error("feeder failed", "err", err)
	}
}
