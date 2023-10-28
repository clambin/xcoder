package ffmpeg

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os/exec"
)

type Processor struct {
	Logger *slog.Logger
}

func (p Processor) Probe(ctx context.Context, path string) (Probe, error) {
	output, err := p.runCommands(ctx,
		"ffprobe",
		"-v", "quiet",
		"-print_format", "json",
		"-show_format", "-show_streams",
		"-print_format", "json",
		path,
	)
	if err != nil {
		return Probe{}, fmt.Errorf("probe: %w", err)
	}

	var probe Probe
	if err = json.NewDecoder(output).Decode(&probe); err != nil {
		err = fmt.Errorf("decode: %w", err)
	}

	return probe, err
}

func (v Processor) Convert(ctx context.Context, input, output, targetCodec string) error {
	if targetCodec == "hevc" {
		targetCodec = "libx265"
	}
	stdout, err := v.runCommands(ctx,
		"ffmpeg",
		"-y",
		"-i", input,
		"-map", "0",
		"-c:v", targetCodec,
		"-c:a", "copy",
		"-c:s", "copy",
		output,
	)
	if err != nil {
		err = fmt.Errorf("failed to convert video. output: %s. err: %w", stdout.String(), err)
	}
	return err
}

func (v Processor) runCommands(ctx context.Context, command string, args ...string) (*bytes.Buffer, error) {
	cmd := exec.CommandContext(ctx, command, args...)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	l := v.Logger.With("cmd", command)

	l.Debug("running command")
	if err := cmd.Run(); err != nil {
		l.Debug("command failed", "err", err)
		return nil, err
	}
	l.Debug("command successful")
	return &stdout, nil
}
