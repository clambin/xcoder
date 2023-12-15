package feeder

import (
	"context"
	"io/fs"
	"log/slog"
	"path/filepath"
	"time"
)

// A Feeder recursively scans a root directory. It looks for files that look like a video and sends those to the Feed channel.
type Feeder struct {
	RootDir  string
	Interval time.Duration
	Feed     chan Entry
	Logger   *slog.Logger
}

type Entry struct {
	Path     string
	DirEntry fs.DirEntry
}

func New(rootDir string, interval time.Duration, logger *slog.Logger) *Feeder {
	return &Feeder{
		RootDir:  rootDir,
		Interval: interval,
		Feed:     make(chan Entry),
		Logger:   logger,
	}
}

func (f Feeder) Run(ctx context.Context) error {
	f.Logger.Debug("started")
	defer f.Logger.Debug("stopped")

	ticker := time.NewTicker(f.Interval)
	defer ticker.Stop()

	for {
		if err := f.scan(ctx); err != nil {
			f.Logger.Error("scan failed", slog.Any("err", err))
		}

		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}
	}
}

var videoExtensions = map[string]struct{}{
	".mkv": {},
	".mp4": {},
	".avi": {},
}

func (f Feeder) scan(ctx context.Context) error {
	err := filepath.WalkDir(f.RootDir, func(path string, d fs.DirEntry, err error) error {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		l := f.Logger.With(slog.String("path", path))
		if err != nil {
			l.Warn("failed to scan directory", "err", err)
			return nil
		}
		if d.IsDir() {
			l.Debug("not a file")
			return nil
		}
		if _, ok := videoExtensions[filepath.Ext(path)]; !ok {
			l.Debug("not a video file")
			return nil
		}

		f.Feed <- Entry{Path: path, DirEntry: d}
		return nil
	})
	return err
}
