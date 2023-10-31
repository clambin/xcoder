package rules

import (
	"github.com/clambin/vidconv/internal/testutil"
	"github.com/clambin/vidconv/internal/video"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestRules_Evaluate(t *testing.T) {
	tests := []struct {
		name       string
		video      video.Video
		wantReason string
		wantOK     bool
	}{
		{
			name:   "valid",
			video:  video.Video{Stats: testutil.MakeProbe("h264", 4800, 1080, time.Hour)},
			wantOK: true,
		},
		{
			name:       "wrong codec",
			video:      video.Video{Stats: testutil.MakeProbe("hevc", 4800, 1080, time.Hour)},
			wantOK:     false,
			wantReason: "video already in hevc",
		},
		{
			name:       "wrong bitrate",
			video:      video.Video{Stats: testutil.MakeProbe("h264", 4000, 1080, time.Hour)},
			wantOK:     false,
			wantReason: "bitrate too low: 4000 kbps",
		},
		{
			name:       "wrong height",
			video:      video.Video{Stats: testutil.MakeProbe("h264", 4800, 720, time.Hour)},
			wantOK:     false,
			wantReason: "height too low: 720",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			reason, ok := MainProfile("hevc").ShouldConvert(tt.video)
			assert.Equal(t, tt.wantOK, ok)
			if !ok {
				assert.Equal(t, tt.wantReason, reason)
			}
		})
	}
}
