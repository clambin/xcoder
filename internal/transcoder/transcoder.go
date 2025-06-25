package transcoder

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os/exec"
	"strconv"

	"github.com/clambin/xcoder/ffmpeg"
)

// Transcoder implements video scanning (ffprobe) and transcoding (ffmpeg)
type Transcoder struct {
	Logger *slog.Logger
}

func (p Transcoder) Scan(_ context.Context, path string) (ffmpeg.VideoStats, error) {
	return ffmpeg.Probe(path)
}

func (p Transcoder) Transcode(ctx context.Context, request Request) error {
	cmd, err := request.buildTranscodeCommand(ctx, request.ProgressCB, p.Logger.With("component", "ffmpeg"))
	if err != nil {
		return fmt.Errorf("failed to create command: %w", err)
	}
	p.Logger.Info("converting", "cmd", cmd.String())
	return cmd.Run()
}

type Request struct {
	ProgressCB  func(ffmpeg.Progress)
	Source      string
	Target      string
	TargetStats ffmpeg.VideoStats
}

var ErrMissingFilename = errors.New("missing filename")
var ErrInvalidCodec = errors.New("unsupported target video codec")
var ErrInvalidBitsPerSample = errors.New("bits per sample must be 8 or 10")
var ErrInvalidBitRate = errors.New("invalid bitrate")

func (r Request) isValid() error {
	if r.Source == "" || r.Target == "" {
		return ErrMissingFilename
	}
	if _, ok := videoCodecs[r.TargetStats.VideoCodec]; !ok {
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

// buildTranscodeCommand creates an exec.Command to run ffmeg for the request
func (r Request) buildTranscodeCommand(ctx context.Context, cb func(ffmpeg.Progress), logger *slog.Logger) (*exec.Cmd, error) {
	if err := r.isValid(); err != nil {
		return nil, err
	}
	videoOutputArgs, err := makeVideoOutputArgs(r.TargetStats)
	if err != nil {
		return nil, err
	}
	videoOutputArgs["c:a"] = "copy"
	videoOutputArgs["c:s"] = "copy"
	videoOutputArgs["f"] = "matroska"

	cmd := ffmpeg.Input(r.Source, inputArguments).
		Output(r.Target, videoOutputArgs).
		NoStats().
		LogLevel("error").
		OverWriteTarget()
	if cb != nil {
		cmd = cmd.WithLogger(logger).ProgressSocket(ctx, cb)
	}
	return cmd.Build(ctx), nil
}

// this may need to move to _darwin/_linux
func makeVideoOutputArgs(stats ffmpeg.VideoStats) (ffmpeg.Args, error) {
	codecName, ok := videoCodecs[stats.VideoCodec]
	if !ok {
		return nil, fmt.Errorf("unsupported video codec: %s", stats.VideoCodec)
	}
	profile := "main"
	if stats.BitsPerSample == 10 {
		profile = "main10"
	}

	return ffmpeg.Args{
		"c:v":       codecName,
		"profile:v": profile,
		"b:v":       strconv.Itoa(stats.BitRate),
	}, nil
}
