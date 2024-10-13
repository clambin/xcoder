package ffmpeg

import (
	"context"
	"fmt"
	ffmpeg "github.com/u2takey/ffmpeg-go"
)

func (p Processor) Scan(_ context.Context, path string) (VideoStats, error) {
	var probe VideoStats

	output, err := ffmpeg.Probe(path)
	if err != nil {
		return probe, fmt.Errorf("probe: %w", err)
	}

	probe, err = Parse(output)
	return probe, nil
}
