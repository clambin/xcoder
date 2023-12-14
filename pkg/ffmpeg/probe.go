package ffmpeg

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
)

func (p Processor) Probe(ctx context.Context, path string) (VideoStats, error) {
	var probe VideoStats

	cmd := exec.CommandContext(ctx,
		"ffprobe",
		//"-v", "quiet",
		"-print_format", "json",
		"-show_format", "-show_streams",
		path,
	)

	stdout, _ := cmd.StdoutPipe()
	defer func() { _ = stdout.Close() }()

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Start()
	if err != nil {
		return probe, fmt.Errorf("start: %w", err)
	}

	if err = json.NewDecoder(stdout).Decode(&probe); err != nil {
		return probe, fmt.Errorf("decode: %w", err)
	}

	if err = cmd.Wait(); err == nil {
		return probe, nil
	}

	return VideoStats{}, fmt.Errorf("probe: %w. error: %s", err, lastLine(&stderr))
}
