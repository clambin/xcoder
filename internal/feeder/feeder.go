package feeder

import (
	"context"
	"errors"
	"fmt"
	"github.com/clambin/vidconv/internal/ffmpeg"
	"io/fs"
	"log/slog"
	"path/filepath"
	"strconv"
	"time"
)

type Feeder struct {
	VideoProcessor
	RootDir  string
	Interval time.Duration
	Feed     chan Video
	Logger   *slog.Logger
}

type VideoProcessor interface {
	Probe(ctx context.Context, path string) (ffmpeg.Probe, error)
}

var _ VideoProcessor = &ffmpeg.Processor{}

func New(rootDir string, interval time.Duration, logger *slog.Logger) *Feeder {
	return &Feeder{
		VideoProcessor: ffmpeg.Processor{Logger: logger.With("component", "ffprobe")},
		RootDir:        rootDir,
		Interval:       interval,
		Feed:           make(chan Video),
		Logger:         logger,
	}
}

func (f Feeder) Run(ctx context.Context) error {
	f.Logger.Info("started")
	defer f.Logger.Info("stopped")
	defer close(f.Feed)
	return f.scan(ctx)
}

func (f Feeder) scan(ctx context.Context) error {
	err := filepath.WalkDir(f.RootDir, func(path string, d fs.DirEntry, err error) error {
		l := f.Logger.With("path", path)
		if err != nil {
			l.Warn("failed to scan directory", "err", err)
			return nil
		}

		video, _ := parseVideoFile(path, d)
		if !video.Info.IsVideo() {
			l.Debug("not a video file")
			return nil
		}

		if video.Duration, video.Codec, err = f.probe(ctx, video.Path); err != nil {
			l.Warn("failed to probe video", "err", err)
			return nil
		}

		f.Logger.Debug("file scanned", "path", path, "codec", video.Codec)
		f.Feed <- video
		return nil
	})
	return err
}

func (f Feeder) probe(ctx context.Context, path string) (time.Duration, string, error) {
	probe, err := f.VideoProcessor.Probe(ctx, path)
	if err != nil {
		return 0, "", fmt.Errorf("probe: %w", err)
	}
	duration, err := strconv.ParseFloat(probe.Format.Duration, 64)
	if err != nil {
		return 0, "", fmt.Errorf("invalud duration '%s': %w", probe.Format.Duration, err)
	}
	codec, err := probe.GetVideoCodec()
	if err != nil {
		return 0, "", errors.New("no video stream found")
	}
	return time.Duration(duration) * time.Second, codec, nil
}
