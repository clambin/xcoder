package pipeline

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/clambin/videoConvertor/internal/configuration"
	"github.com/clambin/videoConvertor/internal/convertor"
	"github.com/clambin/videoConvertor/internal/profile"
)

var (
	ErrAlreadyConverted = errors.New("video already converted")
)

const (
	convertInterval = time.Second
)

func Convert(ctx context.Context, codec Codec, queue *Queue, cfg configuration.Configuration, logger *slog.Logger) {
	convertWithFileChecker(ctx, codec, queue, fsFileChecker{}, cfg, logger)
}

func convertWithFileChecker(ctx context.Context, codec Codec, queue *Queue, f fileChecker, cfg configuration.Configuration, logger *slog.Logger) {
	c := New(codec, queue, cfg, logger)
	c.fileChecker = f
	c.Run(ctx)
}

type Converter struct {
	Codec          Codec
	fileChecker    fileChecker
	List           *Queue
	Logger         *slog.Logger
	Profile        profile.Profile
	RemoveSource   bool
	OverwriteNewer bool
}

type Codec interface {
	Convert(ctx context.Context, request convertor.Request) error
}

type fileChecker interface {
	TargetIsNewer(a, b string) (bool, error)
}

func New(ffmpeg Codec, queue *Queue, cfg configuration.Configuration, l *slog.Logger) *Converter {
	return &Converter{
		Codec:          ffmpeg,
		fileChecker:    fsFileChecker{},
		List:           queue,
		Profile:        cfg.Profile,
		RemoveSource:   cfg.Remove,
		OverwriteNewer: cfg.Overwrite,
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

func (c *Converter) convertItem(ctx context.Context, item *WorkItem) {
	err := c.convert(ctx, item)
	if err == nil {
		c.Logger.Info("converted successfully", "source", item.Source)
		item.SetStatus(Converted, nil)
	} else if errors.Is(err, ErrAlreadyConverted) {
		c.Logger.Info("already converted", "source", item.Source)
		item.SetStatus(Skipped, err)
	} else {
		c.Logger.Warn("conversion failed", "err", err, "source", item.Source)
		item.SetStatus(Failed, err)
	}
}

func (c *Converter) convert(ctx context.Context, item *WorkItem) error {
	// Build target name
	target := buildTargetFilename(item, "", c.Profile.Codec, "mkv")

	// Has the file already been converted?
	targetIsNewer, err := c.fileChecker.TargetIsNewer(item.Source, target)
	if err != nil {
		return err
	}
	if targetIsNewer && !c.OverwriteNewer {
		return ErrAlreadyConverted
	}

	// build the request
	req := convertor.Request{
		Source:      item.Source,
		Target:      target,
		TargetStats: item.TargetVideoStats(),
	}

	cbLogger := c.Logger.With("source", item.Source)
	c.Logger.Info("target determined", "source", item.Source, "bitrate", req.TargetStats.BitRate)

	var lastDurationReported time.Duration
	const reportInterval = 1 * time.Minute
	totalDuration := item.SourceVideoStats().Duration
	req.ProgressCB = func(progress convertor.Progress) {
		completed := progress.Converted.Seconds() / totalDuration.Seconds()
		item.Progress.Update(progress)
		if progress.Converted-lastDurationReported > reportInterval {
			cbLogger.Info("conversion in progress", "progress", progress.Converted, "speed", progress.Speed, "completed", completed)
			lastDurationReported = progress.Converted
		}
	}

	// convert the file
	cbLogger.Debug("converting")
	if err = c.Codec.Convert(ctx, req); err != nil {
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
