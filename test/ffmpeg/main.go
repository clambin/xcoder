package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/clambin/xcoder/ffmpeg"
	"github.com/clambin/xcoder/internal/pipeline"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))
	socketPath := "/tmp/ffmpeg.sock"
	f := ffmpeg.
		Decode(`/Volumes/Video/shows/Law & Order - Organized Crime/Season 5/Law.And.Order.Organized.Crime.s05e01.1080.hevc.mkv`, pipeline.DecoderArguments(ffmpeg.VideoStats{VideoCodec: "hevc"})...).
		Muxer("null").
		NoStats().
		LogLevel("error").
		Progress(func(p ffmpeg.Progress) {
			logger.Debug("progress", "progress", p)
		}, socketPath)
	defer func() { _ = os.Remove(socketPath) }()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	err := f.Run(ctx, logger)
	if err != nil {
		logger.Error(err.Error())
	}
}
