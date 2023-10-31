package ffmpeg

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_getProgress(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		wantProgress Progress
		wantOk       bool
	}{
		{
			name:   "empty",
			input:  "",
			wantOk: false,
		},
		{
			name:   "partial",
			input:  "speed=1.1x\n",
			wantOk: false,
		},
		{
			name:   "partial",
			input:  "out_time_ms=1000\n",
			wantOk: false,
		},
		{
			name:         "valid",
			input:        "speed=1.1x\nout_time_ms=1000\n",
			wantOk:       true,
			wantProgress: Progress{Converted: 1000000, Speed: 1.1},
		},
		{
			name:         "multiple",
			input:        "speed=1.0x\nout_time_ms=1\nspeed=1.1x\nout_time_ms=1000\n",
			wantOk:       true,
			wantProgress: Progress{Converted: 1000000, Speed: 1.1},
		},
		{
			name:         "multiple",
			input:        "speed=1.0x\nout_time_ms=1\nspeed=1.1x\n",
			wantOk:       true,
			wantProgress: Progress{Converted: 1000, Speed: 1.1},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotProgress, gotOk := getProgress(tt.input)
			assert.Equal(t, tt.wantOk, gotOk)
			if tt.wantOk {
				assert.Equal(t, tt.wantProgress, gotProgress)
			}
		})
	}
}
