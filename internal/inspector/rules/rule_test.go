package rules

import (
	"github.com/clambin/vidconv/internal/testutil"
	"github.com/clambin/vidconv/internal/video"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestMinimumHeight(t *testing.T) {
	tests := []struct {
		name       string
		minHeight  int
		video      video.Video
		wantReason string
		wantOK     bool
	}{
		{
			name:      "higher",
			minHeight: 720,
			video:     video.Video{Stats: testutil.MakeProbe("h264", 1000, 1080, time.Hour)},
			wantOK:    true,
		},
		{
			name:      "equal",
			minHeight: 720,
			video:     video.Video{Stats: testutil.MakeProbe("h264", 1000, 720, time.Hour)},
			wantOK:    true,
		},
		{
			name:       "lower",
			minHeight:  720,
			video:      video.Video{Stats: testutil.MakeProbe("h264", 1000, 350, time.Hour)},
			wantOK:     false,
			wantReason: "height too low: 350",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			reason, ok := MinimumHeight(tt.minHeight)(tt.video)
			assert.Equal(t, tt.wantOK, ok)
			if !ok {
				assert.Equal(t, tt.wantReason, reason)
			}
		})
	}
}

func TestOptimumBitrate(t *testing.T) {
	tests := []struct {
		name       string
		video      video.Video
		wantReason string
		wantOK     bool
	}{
		{
			name:   "higher",
			video:  video.Video{Stats: testutil.MakeProbe("h264", 8000, 1080, time.Hour)},
			wantOK: true,
		},
		{
			name:   "equal",
			video:  video.Video{Stats: testutil.MakeProbe("h264", 4800, 1080, time.Hour)},
			wantOK: true,
		},
		{
			name:       "lower",
			video:      video.Video{Stats: testutil.MakeProbe("h264", 80, 1080, time.Hour)},
			wantOK:     false,
			wantReason: "bitrate too low: 80 kbps",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			reason, ok := OptimumBitrate("hevc", 1.2)(tt.video)
			assert.Equal(t, tt.wantOK, ok)
			if !ok {
				assert.Equal(t, tt.wantReason, reason)
			}
		})
	}
}

func TestSkipCodec(t *testing.T) {
	input := video.Video{Stats: testutil.MakeProbe("h264", 4000, 1080, time.Hour)}
	_, ok := SkipCodec("hevc")(input)
	assert.True(t, ok)
	reason, ok := SkipCodec("h264")(input)
	assert.False(t, ok)
	assert.Equal(t, "video already in h264", reason)
}
