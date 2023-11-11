package ffmpeg

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os/exec"
	"strings"
)

func (p Processor) runCommand(ctx context.Context, command string, args ...string) (*bytes.Buffer, error) {
	cmd := exec.CommandContext(ctx, command, args...)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	l := p.Logger.With(slog.String("cmd", command))

	l.Debug("running command")
	if err := cmd.Run(); err != nil {
		lines := strings.Split(stderr.String(), "\n")
		var lastLine string
		if len(lines) > 0 {
			lastLine = lines[len(lines)-1]
		}
		return nil, fmt.Errorf("command failed: %w. last line: %s", err, lastLine)
	}
	l.Debug("command successful")
	return &stdout, nil
}
