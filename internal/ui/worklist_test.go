package ui

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func Test_formatDuration(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Duration
		expected string
	}{
		{"No duration", 0, ""},
		{"Only seconds", 10 * time.Second, "10s"},
		{"Only minutes", 5 * time.Minute, "5m"},
		{"Only hours", 2 * time.Hour, "2h"},
		{"Days and hours", 26 * time.Hour, "1d2h"},
		{"Hours and minutes", 3*time.Hour + 45*time.Minute, "3h45m"},
		{"Hours, minutes, seconds", 1*time.Hour + 30*time.Minute + 20*time.Second, "1h30m20s"},
		{"Hours and seconds", 1*time.Hour + 2*time.Second, "1h2s"},
		{"Exactly two days", 48 * time.Hour, "2d"},
		{"Multiple units", 72*time.Hour + 10*time.Minute + 15*time.Second, "3d10m15s"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, formatDuration(tt.input))
		})
	}
}
