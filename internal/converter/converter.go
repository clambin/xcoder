package converter

import (
	"context"
	"errors"
	"fmt"
	"github.com/clambin/videoConvertor/internal/configuration"
	"github.com/clambin/videoConvertor/internal/ffmpeg"
	"github.com/clambin/videoConvertor/internal/profile"
	"github.com/clambin/videoConvertor/internal/worklist"
	"log/slog"
	"os"
	"time"
)

var (
	ErrAlreadyConverted = errors.New("video already converted")
)

const (
	convertInterval = time.Second
)

type Converter struct {
	FFMPEG
	fileChecker
	List           *worklist.WorkList
	Logger         *slog.Logger
	Profile        profile.Profile
	RemoveSource   bool
	OverwriteNewer bool
}

type FFMPEG interface {
	Convert(ctx context.Context, request ffmpeg.Request) error
}

type fileChecker interface {
	TargetIsNewer(a, b string) (bool, error)
}

func New(ffmpeg FFMPEG, w *worklist.WorkList, cfg configuration.Configuration, l *slog.Logger) *Converter {
	return &Converter{
		FFMPEG:         ffmpeg,
		fileChecker:    fsFileChecker{},
		List:           w,
		Profile:        cfg.Profile,
		RemoveSource:   cfg.RemoveSource,
		OverwriteNewer: cfg.OverwriteNewerTarget,
		Logger:         l,
	}
}

func (c *Converter) Run(ctx context.Context) {
	ticker := time.NewTicker(convertInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		if item := c.List.NextToConvert(); item != nil {
			c.convertItem(ctx, item)
			continue
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (c *Converter) convertItem(ctx context.Context, item *worklist.WorkItem) {
	err := c.convert(ctx, item)
	if err == nil {
		c.Logger.Info("converted successfully", "source", item.Source)
		item.Done(worklist.Converted, nil)
	} else if errors.Is(err, ErrAlreadyConverted) {
		c.Logger.Info("already converted", "source", item.Source)
		item.Done(worklist.Skipped, err)
	} else {
		c.Logger.Warn("conversion failed", "err", err, "source", item.Source)
		item.Done(worklist.Failed, err)
	}
}

func (c *Converter) convert(ctx context.Context, item *worklist.WorkItem) error {
	// Build target name
	target := makeTargetFilename(item, "", c.Profile.Codec, "mkv")

	// Has the file already been converted?
	targetIsNewer, err := c.fileChecker.TargetIsNewer(item.Source, target)
	if err != nil {
		return err
	}
	if targetIsNewer && !c.OverwriteNewer {
		return ErrAlreadyConverted
	}

	// build the request
	req := ffmpeg.Request{
		Source:        item.Source,
		Target:        target,
		VideoCodec:    item.TargetVideoStats().VideoCodec(),
		BitsPerSample: item.SourceVideoStats().BitsPerSample(),
		BitRate:       item.TargetVideoStats().BitRate(),
	}

	cbLogger := c.Logger.With("source", item.Source)
	c.Logger.Info("target determined", "source", item.Source, "bitrate", req.BitRate)

	var lastDurationReported time.Duration
	const reportInterval = 1 * time.Minute
	totalDuration := item.SourceVideoStats().Duration().Seconds()
	req.ProgressCB = func(progress ffmpeg.Progress) {
		completed := progress.Converted.Seconds() / totalDuration
		item.Progress.Update(progress)
		if progress.Converted-lastDurationReported > reportInterval {
			cbLogger.Info("conversion in progress", "progress", progress.Converted, "speed", progress.Speed, "completed", completed)
			lastDurationReported = progress.Converted
		}
	}

	// convert the file
	cbLogger.Debug("converting")
	if err = c.FFMPEG.Convert(ctx, req); err != nil {
		_ = os.Remove(target)
		return fmt.Errorf("failed to convert video: %w", err)
	}

	// clean up if needed
	if c.RemoveSource {
		if err = os.Remove(item.Source); err != nil {
			c.Logger.Warn("failed to remove source file", "err", err, "source", item.Source)
		}
	}
	return nil
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
