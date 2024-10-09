package ffmpeg

import (
	"context"
	"encoding/json"
	"fmt"
	ffmpeg "github.com/u2takey/ffmpeg-go"
)

func (p Processor) Scan(_ context.Context, path string) (VideoStats, error) {
	var probe VideoStats

	output, err := ffmpeg.Probe(path)
	if err != nil {
		return probe, fmt.Errorf("probe: %w", err)
	}

	if err = json.Unmarshal([]byte(output), &probe); err != nil {
		return probe, fmt.Errorf("decode: %w", err)
	}

	return probe, nil
}
