package quality

import (
	"github.com/clambin/videoConvertor/internal/testutil"
	"github.com/clambin/videoConvertor/pkg/ffmpeg"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func Test_getMinimumBitrate(t *testing.T) {
	type args struct {
		codec  string
		height int
	}
	testCases := []struct {
		name   string
		args   args
		wantOK assert.BoolAssertionFunc
		want   int
	}{
		{
			name:   "h264 480",
			args:   args{codec: "h264", height: 480},
			wantOK: assert.True,
			want:   1500 * 1024,
		},
		{
			name:   "hevc 480",
			args:   args{codec: "hevc", height: 480},
			wantOK: assert.True,
			want:   750 * 1024,
		},
		{
			name:   "hevc 719",
			args:   args{codec: "hevc", height: 719},
			wantOK: assert.True,
			want:   1500 * 1024,
		},
		{
			name:   "hevc 720",
			args:   args{codec: "hevc", height: 720},
			wantOK: assert.True,
			want:   1500 * 1024,
		},
		{
			name:   "hevc 721",
			args:   args{codec: "hevc", height: 721},
			wantOK: assert.True,
			want:   1500 * 1024,
		},
		{
			name:   "hevc 900",
			args:   args{codec: "hevc", height: 900},
			wantOK: assert.True,
			want:   1500 * 1024,
		},
		{
			name:   "hevc 901",
			args:   args{codec: "hevc", height: 901},
			wantOK: assert.True,
			want:   3 * 1024 * 1024,
		},
		{
			name:   "invalid codec",
			args:   args{codec: "invalid", height: 480},
			wantOK: assert.False,
		},
	}

	for _, tt := range testCases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			bitrate, ok := getMinimumBitrate(tt.args.codec, tt.args.height)
			tt.wantOK(t, ok)
			assert.Equal(t, tt.want, bitrate)
		})
	}
}

func Test_GetMinimumBitrate(t *testing.T) {
	tt := []struct {
		name   string
		stats  ffmpeg.VideoStats
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
			want:   3 * 1024 * 1024,
		},
		{
			name:   "2000",
			stats:  testutil.MakeProbe("hevc", 1000, 2000, time.Hour),
			wantOK: assert.True,
			want:   15 * 1024 * 1024,
		},
		{
			name:   "4000",
			stats:  testutil.MakeProbe("hevc", 1000, 4000, time.Hour),
			wantOK: assert.True,
			want:   15 * 1024 * 1024,
		},
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			bitRate, ok := GetMinimumBitRate(tc.stats, "hevc")
			tc.wantOK(t, ok)
			if ok {
				assert.Equal(t, tc.want, bitRate)
			}
		})
	}
}
