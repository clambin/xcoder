package main

import (
	"context"
	"flag"
	"github.com/clambin/videoConvertor/internal/server"
	"github.com/clambin/videoConvertor/internal/server/scanner"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

var (
	debug   = flag.Bool("debug", false, "switch on debug logging")
	input   = flag.String("input", "/media", "input directory")
	profile = flag.String("profile", "hevc-max", "conversion profile")
	active  = flag.Bool("active", false, "start convertor in active mode")
	remove  = flag.Bool("remove", false, "remove source files after successful conversion")
	//concurrent = flag.Int("concurrent", 1, "number of concurrent video convertors ")
)

func main() {
	flag.Parse()

	var handlerOpts slog.HandlerOptions
	if *debug {
		handlerOpts.Level = slog.LevelDebug
	}

	l := slog.New(slog.NewTextHandler(os.Stderr, &handlerOpts))

	cfg := server.Config{
		Addr: ":9090",
		ScannerConfig: scanner.Config{
			RootDir: *input,
			Profile: *profile,
		},
		RemoveConverted: *remove,
	}
	s, err := server.New(cfg, l)
	if err != nil {
		l.Error("failed to create server", "err", err)
	}
	s.Convertor.SetActive(*active)
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err = s.Run(ctx); err != nil {
		l.Error("feeder failed", "err", err)
	}
}
