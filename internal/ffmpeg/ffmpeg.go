package ffmpeg

import (
	"context"
	"errors"
	"fmt"
	ffmpeg "github.com/u2takey/ffmpeg-go"
	"log/slog"
)

type Request struct {
	ProgressCB  func(Progress)
	Source      string
	Target      string
	TargetStats VideoStats
}

var ErrMissingFilename = errors.New("missing filename")
var ErrInvalidCodec = errors.New("only hevc supported")
var ErrInvalidBitsPerSample = errors.New("bits per sample must be 8 or 10")
var ErrInvalidBitRate = errors.New("invalid bitrate")

func (r Request) IsValid() error {
	if r.Source == "" || r.Target == "" {
		return ErrMissingFilename
	}
	if r.TargetStats.VideoCodec != "hevc" {
		return ErrInvalidCodec
	}
	if r.TargetStats.BitsPerSample != 8 && r.TargetStats.BitsPerSample != 10 {
		return ErrInvalidBitsPerSample
	}
	if r.TargetStats.BitRate == 0 {
		return ErrInvalidBitRate
	}
	return nil
}

// Processor implements video scanning (ffprobe) and converting (ffmpeg).  Implemented as a struct so that it can be
// mocked at the calling side.
type Processor struct {
	Logger *slog.Logger
}

func (p Processor) Scan(_ context.Context, path string) (VideoStats, error) {
	var probe VideoStats

	output, err := ffmpeg.Probe(path)
	if err != nil {
		return probe, fmt.Errorf("probe: %w", err)
	}

	return parseVideoStats(output)
}

func (p Processor) Convert(ctx context.Context, request Request) error {
	if err := request.IsValid(); err != nil {
		return err
	}

	var sock string
	if request.ProgressCB != nil {
		var err error
		if sock, err = p.progressSocket(request.ProgressCB); err != nil {
			return fmt.Errorf("progress socket: %w", err)
		}
	}
	stream, err := makeConvertCommand(ctx, request, sock)
	if err != nil {
		return fmt.Errorf("failed to create command: %w", err)
	}
	p.Logger.Info("converting", "cmd", stream.Compile().String())
	return stream.Run()
}
