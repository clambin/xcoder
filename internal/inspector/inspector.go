package inspector

import (
	"context"
	"errors"
	"fmt"
	"github.com/clambin/vidconv/internal/feeder"
	"github.com/clambin/vidconv/pkg/ffmpeg"
	"log/slog"
)

type Inspector struct {
	VideoProcessor
	targetCodec string
	input       <-chan feeder.Entry
	Output      chan Video
	logger      *slog.Logger
}

type VideoProcessor interface {
	Probe(ctx context.Context, path string) (ffmpeg.Probe, error)
}

func New(input <-chan feeder.Entry, targetCodec string, logger *slog.Logger) *Inspector {
	return &Inspector{
		VideoProcessor: ffmpeg.Processor{Logger: logger.With("component", "ffprobe")},
		targetCodec:    targetCodec,
		input:          input,
		Output:         make(chan Video),
		logger:         logger,
	}
}

func (i Inspector) Run(ctx context.Context) error {
	i.logger.Info("started")
	defer i.logger.Info("stopped")
	for {
		select {
		case entry := <-i.input:
			i.logger.Debug("probe", "path", entry.Path)
			video, err := i.scan(ctx, entry)
			if err != nil {
				i.logger.Error("failed to get video stats", "err", err)
				break
			}
			if video.Stats.VideoCodec() == i.targetCodec {
				i.logger.Debug("video already in target codec", "path", entry.Path)
				break
			}
			i.Output <- video
		case <-ctx.Done():
			return nil
		}
	}
}

func (i Inspector) scan(ctx context.Context, entry feeder.Entry) (Video, error) {
	video := Video{Path: entry.Path}
	if entry.DirEntry.IsDir() {
		return video, errors.New("is a directory")
	}

	fileInfo, err := entry.DirEntry.Info()
	if err != nil {
		return video, fmt.Errorf("file info: %w", err)
	}
	video.ModTime = fileInfo.ModTime()

	var ok bool
	if video.Info, ok = parseVideoFilename(entry.DirEntry.Name()); !ok {
		return video, errors.New("failed to parse filename")
	}

	video.Stats, err = i.VideoProcessor.Probe(ctx, entry.Path)
	if err != nil {
		return video, fmt.Errorf("probe: %w", err)
	}

	return video, nil
}
