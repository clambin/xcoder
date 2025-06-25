package pipeline

import (
	"context"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/clambin/xcoder/ffmpeg"
	"github.com/clambin/xcoder/internal/configuration"
	"github.com/clambin/xcoder/internal/transcoder"
)

const (
	convertInterval = 100 * time.Millisecond
)

func Transcode(ctx context.Context, tc Transcoder, queue *Queue, cfg configuration.Configuration, logger *slog.Logger) {
	transcodeWithFileChecker(ctx, tc, queue, fsFileChecker{}, cfg, logger)
}

type Transcoder interface {
	Transcode(ctx context.Context, request transcoder.Request) error
}

type fileChecker interface {
	TargetIsNewer(a, b string) (bool, error)
}

func transcodeWithFileChecker(ctx context.Context, tc Transcoder, queue *Queue, f fileChecker, cfg configuration.Configuration, logger *slog.Logger) {
	ticker := time.NewTicker(convertInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
		if item := queue.NextToConvert(); item != nil {
			transcodeItem(ctx, item, tc, f, cfg, logger)
		}
	}
}

func transcodeItem(ctx context.Context, item *WorkItem, tc Transcoder, f fileChecker, cfg configuration.Configuration, logger *slog.Logger) {
	// Build target name
	target := buildTargetFilename(item, "", cfg.Profile.TargetCodec, "mkv")
	logger = logger.With("target", target)

	// Has the file already been converted?
	targetIsNewer, err := f.TargetIsNewer(item.Source, target)
	if err != nil {
		logger.Error("failed to check if target is newer", "err", err)
		item.SetStatus(Failed, err)
	}
	if targetIsNewer && !cfg.Overwrite {
		logger.Info("already converted")
		item.SetStatus(Skipped, err)
		return
	}

	// build the request
	var lastDurationReported time.Duration
	const reportInterval = 1 * time.Minute
	totalDuration := item.SourceVideoStats().Duration
	req := transcoder.Request{
		Source:      item.Source,
		Target:      target,
		TargetStats: item.TargetVideoStats(),
		ProgressCB: func(progress ffmpeg.Progress) {
			item.Progress.Update(progress)
			if progress.Converted-lastDurationReported > reportInterval {
				logger.Info("conversion in progress",
					"progress", progress.Converted,
					"speed", progress.Speed,
					"completed", strconv.FormatFloat(100*progress.Converted.Seconds()/totalDuration.Seconds(), 'f', 2, 64)+"%",
				)
				lastDurationReported = progress.Converted
			}
		},
	}

	logger.Info("target determined", "bitrate", req.TargetStats.BitRate)

	// convert the file
	logger.Debug("converting")
	if err = tc.Transcode(ctx, req); err != nil {
		_ = os.Remove(target)
		logger.Warn("conversion failed", "err", err)
		item.SetStatus(Failed, err)
		return
	}

	// clean up if needed
	if cfg.Remove {
		if err = os.Remove(item.Source); err != nil {
			logger.Warn("failed to remove source file", "err", err)
			item.SetStatus(Failed, err)
		}
	}
	logger.Info("converted successfully")
	item.SetStatus(Converted, nil)
}

type fsFileChecker struct{}

// TargetIsNewer returns true if target exists and is more recent than target. If there was an error checking source, error contains the error.
// If there was an error checking target, it returns (true, nil).
func (c fsFileChecker) TargetIsNewer(source, target string) (bool, error) {
	sStats, err := os.Stat(source)
	if err != nil {
		return false, err
	}
	tStats, err := os.Stat(target)
	if err != nil {
		return false, nil
	}
	return tStats.ModTime().After(sStats.ModTime()), nil
}
