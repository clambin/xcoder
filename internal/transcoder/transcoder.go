package transcoder

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os/exec"
	"strconv"

	"github.com/clambin/videoConvertor/ffmpeg"
)

// Transcoder implements video scanning (ffprobe) and transcoding (ffmpeg)
type Transcoder struct {
	Logger *slog.Logger
}

func (p Transcoder) Scan(_ context.Context, path string) (ffmpeg.VideoStats, error) {
	return ffmpeg.Probe(path)
}

func (p Transcoder) Transcode(ctx context.Context, request Request) error {
	if err := request.IsValid(); err != nil {
		return err
	}
	cmd, err := makeTranscodeCommand(ctx, request, request.ProgressCB, p.Logger.With("component", "ffmpeg"))
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

// makeTranscodeCommand creates an exec.Command to run ffmeg with the required configuration.
func makeTranscodeCommand(ctx context.Context, request Request, cb func(progress ffmpeg.Progress), logger *slog.Logger) (*exec.Cmd, error) {
	codecName, ok := videoCodecs[request.TargetStats.VideoCodec]
	if !ok {
		return nil, fmt.Errorf("unsupported video codec: %s", request.TargetStats.VideoCodec)
	}
	profile := "main"
	if request.TargetStats.BitsPerSample == 10 {
		profile = "main10"
	}

	cmd := ffmpeg.Input(request.Source, inputArguments).
		Output(request.Target, ffmpeg.Args{
			"c:v":       codecName,
			"profile:v": profile,
			"b:v":       strconv.Itoa(request.TargetStats.BitRate),
			"c:a":       "copy",
			"c:s":       "copy",
			"f":         "matroska",
		}).
		NoStats().
		LogLevel("error").
		OverWriteTarget()
	if cb != nil {
		cmd = cmd.WithLogger(logger).ProgressSocket(ctx, cb)
	}
	return cmd.Build(ctx), nil
}
