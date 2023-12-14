package inspector

import (
	"context"
	"errors"
	"fmt"
	"github.com/clambin/vidconv/internal/server/requests"
	"github.com/clambin/vidconv/internal/server/scanner/feeder"
	"github.com/clambin/vidconv/internal/server/scanner/inspector/quality"
	"github.com/clambin/vidconv/pkg/ffmpeg"
	"log/slog"
	"os"
)

// An Inspector receives files (from a Feeder).  It scans the file (using ffprobe) to determine the video characteristics of the file.
// The scanned file is then sent to the Target channel for conversion.
type Inspector struct {
	VideoProcessor
	profile  quality.Profile
	input    <-chan feeder.Entry
	requests *requests.Requests
	logger   *slog.Logger
}

type VideoProcessor interface {
	Probe(ctx context.Context, path string) (ffmpeg.VideoStats, error)
}

func New(input <-chan feeder.Entry, profile string, requests *requests.Requests, logger *slog.Logger) *Inspector {
	p, err := quality.GetProfile(profile)
	if err != nil {
		// TODO
		panic("invalid profile")
	}
	return &Inspector{
		VideoProcessor: ffmpeg.Processor{Logger: logger.With(slog.String("component", "ffprobe"))},
		profile:        p,
		input:          input,
		requests:       requests,
		logger:         logger,
	}
}

func (i Inspector) Run(ctx context.Context) error {
	i.logger.Debug("started")
	defer i.logger.Debug("stopped")
	for {
		select {
		case entry := <-i.input:
			if req, err := i.process(ctx, entry); err == nil {
				i.requests.Add(req)
				i.logger.Info("queued", slog.Any("request", req))
			} else if errors.Is(err, quality.ErrSourceRejected{}) {
				i.logger.Debug("skipped", slog.String("source", entry.Path), slog.Any("reason", err))
			} else {
				i.logger.Error("failed to process", slog.String("source", entry.Path), slog.Any("err", err))
			}
		case <-ctx.Done():
			return nil
		}
	}
}

func (i Inspector) process(ctx context.Context, entry feeder.Entry) (requests.Request, error) {
	target := makeTargetFilename(entry.Path, "", i.profile.Codec, "mkv")

	converted, err := alreadyConverted(entry.Path, target)
	if err != nil {
		return requests.Request{}, fmt.Errorf("alreadyConverted: %w", err)
	}
	if converted {
		return requests.Request{}, quality.ErrSourceRejected{Reason: "already converted"}
	}

	sourceStats, err := i.VideoProcessor.Probe(ctx, entry.Path)
	if err != nil {
		return requests.Request{}, fmt.Errorf("probe source: %w", err)
	}

	return i.profile.MakeRequest(target, entry.Path, sourceStats)
}

func alreadyConverted(source, target string) (bool, error) {
	sourceInfo, err := os.Stat(source)
	if err != nil {
		return false, fmt.Errorf("source stat: %w", err)
	}
	targetInfo, err := os.Stat(target)
	return err == nil && targetInfo.ModTime().After(sourceInfo.ModTime()), nil
}
