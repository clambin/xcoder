package video

import (
	"github.com/clambin/vidconv/internal/testutil"
	"github.com/clambin/vidconv/pkg/ffmpeg"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func Test_getMinimumFrameRate(t *testing.T) {
	tt := []struct {
		name   string
		stats  ffmpeg.Probe
		wantOK assert.BoolAssertionFunc
		want   int
	}{
		{
			name:   "480",
			stats:  testutil.MakeProbe("hevc", 1000, 480, time.Hour),
			wantOK: assert.True,
			want:   750 * 1024,
		},
		{
			name:   "720",
			stats:  testutil.MakeProbe("hevc", 1000, 720, time.Hour),
			wantOK: assert.True,
			want:   1500 * 1024,
		},
		{
			name:   "1080",
			stats:  testutil.MakeProbe("hevc", 1000, 1080, time.Hour),
			wantOK: assert.True,
			want:   4000 * 1024,
		},
		{
			name:   "2000",
			stats:  testutil.MakeProbe("hevc", 1000, 2000, time.Hour),
			wantOK: assert.True,
			want:   15000 * 1024,
		},
		{
			name:   "4000",
			stats:  testutil.MakeProbe("hevc", 1000, 4000, time.Hour),
			wantOK: assert.True,
			want:   15000 * 1024,
		},
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			bitRate, ok := GetMinimumBitRate(Video{Stats: tc.stats}, "hevc")
			tc.wantOK(t, ok)
			if ok {
				assert.Equal(t, tc.want, bitRate)
			}
		})
	}
}
