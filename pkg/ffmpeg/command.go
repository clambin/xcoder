package ffmpeg

import (
	"bytes"
	"context"
	"os/exec"
	"strings"
)

func (p Processor) runCommand(ctx context.Context, command string, args ...string) (*bytes.Buffer, *bytes.Buffer, error) {
	cmd := exec.CommandContext(ctx, command, args...)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	return &stdout, &stderr, cmd.Run()
}

func lastLine(buffer *bytes.Buffer) string {
	lines := strings.Split(buffer.String(), "\n")
	for count := len(lines) - 1; count >= 0; count-- {
		if lines[count] != "" {
			return lines[count]
		}
	}
	return ""
}
