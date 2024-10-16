package ffmpeg

import (
	"context"
	"fmt"
	"github.com/clambin/videoConvertor/internal/ffmpeg/cmd"
	"log/slog"
	"net"
)

type Request struct {
	ProgressCB         func(Progress)
	Source             string
	SourceStats        VideoStats
	Target             string
	TargetVideoCodec   string
	ConstantRateFactor int
}

func (r Request) IsValid() error {
	if r.Source == "" || r.Target == "" {
		return ErrMissingFilename
	}
	//if r.SourceStats.Height == 0 {
	//	return ErrMissingHeight
	//}
	if r.SourceStats.BitsPerSample != 8 && r.SourceStats.BitsPerSample != 10 {
		return ErrInvalidBitsPerSample
	}
	if r.TargetVideoCodec != "hevc" {
		return ErrInvalidCodec
	}
	if r.ConstantRateFactor <= 0 || r.ConstantRateFactor > 51 {
		return ErrInvalidConstantRateFactor{ConstantRateFactor: r.ConstantRateFactor}
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

	output, err := cmd.Probe(path)
	if err != nil {
		return probe, fmt.Errorf("probe: %w", err)
	}

	return parseVideoStats(output)
}

func (p Processor) Convert(ctx context.Context, request Request) error {
	if err := request.IsValid(); err != nil {
		return err
	}

	var progressSocketPath string
	if request.ProgressCB != nil {
		var progressSocketListener net.Listener
		var err error
		progressSocketListener, progressSocketPath, err = p.makeProgressSocket()
		if err != nil {
			return fmt.Errorf("progress socket: %w", err)
		}
		go p.serveProgressSocket(progressSocketListener, progressSocketPath, request.ProgressCB)

	}
	cmd, err := makeConvertCommand(ctx, request, progressSocketPath)
	if err != nil {
		return fmt.Errorf("failed to create command: %w", err)
	}
	p.Logger.Info("converting", "cmd", cmd.String())
	return cmd.Run()
}
