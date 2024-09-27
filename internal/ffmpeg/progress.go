package ffmpeg

import (
	"strconv"
	"strings"
	"time"
)

type Progress struct {
	Converted time.Duration
	Speed     float64
}

func getProgress(input string) (Progress, bool) {
	var foundSpeed, foundConverted bool
	var progress Progress
	lines := strings.Split(input, "\n")

	for i := len(lines) - 1; !(foundSpeed && foundConverted) && i >= 0; i-- {
		if !foundSpeed {
			progress.Speed, foundSpeed = getSpeed(lines[i])
		}
		if !foundConverted {
			progress.Converted, foundConverted = getConverted(lines[i])
		}
	}
	return progress, foundSpeed && foundConverted
}

func getSpeed(input string) (float64, bool) {
	const prefix = "speed="
	var speed float64
	if strings.HasPrefix(input, prefix) {
		line := strings.TrimPrefix(input, prefix)
		line = strings.TrimSuffix(line, "x")
		speed, _ = strconv.ParseFloat(line, 64)
	}
	return speed, speed > 0
}

func getConverted(input string) (time.Duration, bool) {
	const prefix = "out_time_ms="
	var converted int
	if strings.HasPrefix(input, prefix) {
		line := strings.TrimPrefix(input, prefix)
		converted, _ = strconv.Atoi(line)
	}
	return time.Duration(converted) * time.Microsecond, converted != 0
}
