package ffmpeg

import (
	"regexp"
	"strconv"
	"time"
)

func getProgress(input string) (progress Progress, ok bool) {
	if progress.Converted, ok = getLastProgress(input); !ok {
		return progress, ok
	}
	progress.Speed, ok = getLastSpeed(input)
	return progress, ok
}

var (
	regexpOutTime = regexp.MustCompile(`out_time_ms=(\d+)\n`)
	regexpSpeed   = regexp.MustCompile(`speed=(\d+\.\d+)x`)
)

func getLastProgress(input string) (time.Duration, bool) {
	value, ok := parseProgress(regexpOutTime, input)
	if !ok {
		return 0, false
	}
	progress, _ := strconv.Atoi(value)
	return time.Duration(progress) * time.Microsecond, true
}

func getLastSpeed(input string) (float64, bool) {
	value, ok := parseProgress(regexpSpeed, input)
	if !ok {
		return 0, false
	}
	speed, _ := strconv.ParseFloat(value, 64)
	return speed, true
}

func parseProgress(re *regexp.Regexp, input string) (string, bool) {
	matches := re.FindAllStringSubmatch(input, -1)
	if len(matches) == 0 {
		return "", false
	}
	count1 := len(matches)
	count2 := len(matches[count1-1])
	return matches[count1-1][count2-1], true
}
