package inspector

import (
	"context"
	"errors"
	"fmt"
	"github.com/clambin/vidconv/internal/feeder"
	"github.com/clambin/vidconv/internal/inspector/rules"
	"github.com/clambin/vidconv/internal/video"
	"github.com/clambin/vidconv/pkg/ffmpeg"
	"log/slog"
	"os"
)

// An Inspector receives files (from a Feeder).  It scans the file (using ffprobe) to determine the video characteristics of the file.
// The scanned file is then sent to the Output channel for conversion.
type Inspector struct {
	rules rules.Rules
	VideoProcessor
	targetCodec string
	input       <-chan feeder.Entry
	Output      chan ConversionRequest
	logger      *slog.Logger
}

type VideoProcessor interface {
	Probe(ctx context.Context, path string) (ffmpeg.Probe, error)
}

type ConversionRequest struct {
	Input         video.Video
	TargetFile    string
	TargetCodec   string
	TargetBitrate int
}

func New(input <-chan feeder.Entry, targetCodec string, logger *slog.Logger) *Inspector {
	return &Inspector{
		VideoProcessor: ffmpeg.Processor{Logger: logger.With(slog.String("component", "ffprobe"))},
		targetCodec:    targetCodec,
		rules:          rules.MainProfile(targetCodec),
		input:          input,
		Output:         make(chan ConversionRequest),
		logger:         logger,
	}
}

func (i Inspector) Run(ctx context.Context) error {
	i.logger.Info("started")
	defer i.logger.Info("stopped")
	for {
		select {
		case entry := <-i.input:
			l := i.logger.With(slog.String("path", entry.Path))

			input, err := i.scan(ctx, entry)
			if err != nil {
				l.Error("failed to scan video", slog.Any("err", err))
				break
			}

			request, reason, convert := i.makeRequest(input)
			if !convert {
				logLvl := slog.LevelWarn
				if reason == "video already in hevc" {
					logLvl = slog.LevelDebug
				}
				l.Log(ctx, logLvl, "skipping video", slog.String("reason", reason))

				break
			}
			i.Output <- request
		case <-ctx.Done():
			return nil
		}
	}
}

func (i Inspector) makeRequest(input video.Video) (ConversionRequest, string, bool) {
	if reason, ok := i.rules.ShouldConvert(input); !ok {
		return ConversionRequest{}, reason, false
	}

	bitrate, _ := video.GetMinimumBitRate(input, i.targetCodec)

	req := ConversionRequest{
		Input:         input,
		TargetFile:    input.NameWithCodec(i.targetCodec),
		TargetCodec:   i.targetCodec,
		TargetBitrate: bitrate,
	}

	outfileInfo, err := os.Stat(req.TargetFile)
	if err == nil && outfileInfo.ModTime().After(req.Input.ModTime) {
		return req, "file already converted", false
	}

	return req, "", true
}

func (i Inspector) scan(ctx context.Context, entry feeder.Entry) (video.Video, error) {
	input := video.Video{Path: entry.Path}
	if entry.DirEntry.IsDir() {
		return input, errors.New("is a directory")
	}

	fileInfo, err := entry.DirEntry.Info()
	if err != nil {
		return input, fmt.Errorf("file info: %w", err)
	}
	input.ModTime = fileInfo.ModTime()

	var ok bool
	if input.Info, ok = video.ParseVideoFilename(entry.DirEntry.Name()); !ok {
		return input, errors.New("failed to parse filename")
	}

	input.Stats, err = i.VideoProcessor.Probe(ctx, entry.Path)
	if err != nil {
		return input, fmt.Errorf("probe: %w", err)
	}

	return input, nil
}
