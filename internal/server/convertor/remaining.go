package convertor

import "time"

func calculateRemaining(start, now time.Time, converted, total time.Duration) time.Duration {
	elapsed := now.Sub(start)
	speed := float64(converted) / float64(elapsed)

	remaining := total - converted

	return time.Duration(float64(remaining) / speed)
}
