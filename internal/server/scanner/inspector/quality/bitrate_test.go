package quality

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
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

func TestBitrateRatio(t *testing.T) {
	t.Skip()
	for _, height := range []int{480, 720, 1080, 2160} {
		hvec, ok := getMinimumBitrate("hevc", height)
		require.True(t, ok)
		h264, ok := getMinimumBitrate("h264", height)
		require.True(t, ok)

		t.Log(height, float64(h264)/float64(hvec))
	}
}
