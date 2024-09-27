package preprocessor

import (
	"context"
	"errors"
	"fmt"
	"github.com/clambin/videoConvertor/internal/ffmpeg"
	"github.com/clambin/videoConvertor/internal/profile"
	"github.com/clambin/videoConvertor/internal/worklist"
	"log/slog"
	"time"
)

type FFMPEG interface {
	Scan(ctx context.Context, path string) (ffmpeg.VideoStats, error)
}

const scanInterval = time.Second

func Run(ctx context.Context, ch <-chan *worklist.WorkItem, prober FFMPEG, profile profile.Profile, logger *slog.Logger) {
	ticker := time.NewTicker(scanInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case item := <-ch:
			l := logger.With("source", item.Source)
			if err := preProcess(ctx, item, prober, profile, l); err != nil {
				l.Warn("preprocessing failed", "err", err)
			}
		}
	}
}

func preProcess(ctx context.Context, item *worklist.WorkItem, codec FFMPEG, videoProfile profile.Profile, l *slog.Logger) error {
	l.Debug("preprocessing file")
	sourceStats, err := codec.Scan(ctx, item.Source)
	if err != nil {
		l.Warn("failed to preprocess file", "err", err)
		err = fmt.Errorf("preprocessing failed: %w", err)
		item.Done(worklist.Failed, err)
		return err
	}
	item.AddSourceStats(sourceStats)
	l.Debug("preprocessing done", "sourceStats", sourceStats)

	// Validate that the video meets the criteria
	targetStats, err := videoProfile.Evaluate(sourceStats)
	if err != nil {
		l.Debug("should not convert file", "err", err)
		status := worklist.Rejected
		if errors.Is(err, profile.ErrSourceInTargetCodec) {
			status = worklist.Skipped
		}
		item.Done(status, err)
		return err
	}
	item.AddTargetStats(targetStats)

	// Add the sourceStats
	item.Done(worklist.Inspected, nil)
	return nil
}
