package quality

import (
	"github.com/clambin/videoConvertor/internal/testutil"
	"github.com/clambin/videoConvertor/pkg/ffmpeg"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestSkipCodec(t *testing.T) {
	stats := testutil.MakeProbe("h264", 4000, 1080, time.Hour)
	assert.NoError(t, SkipCodec("hevc")(stats))
	err := SkipCodec("h264")(stats)
	assert.Error(t, err)
	assert.Equal(t, "source rejected: video already in target codec", err.Error())
}

func TestMinimumHeight(t *testing.T) {
	tests := []struct {
		name      string
		minHeight int
		stats     ffmpeg.VideoStats
		wantErr   assert.ErrorAssertionFunc
		want      string
	}{
		{
			name:      "higher",
			minHeight: 720,
			stats:     testutil.MakeProbe("h264", 1000, 1080, time.Hour),
			wantErr:   assert.NoError,
		},
		{
			name:      "equal",
			minHeight: 720,
			stats:     testutil.MakeProbe("h264", 1000, 720, time.Hour),
			wantErr:   assert.NoError,
		},
		{
			name:      "lower",
			minHeight: 720,
			stats:     testutil.MakeProbe("h264", 1000, 350, time.Hour),
			wantErr:   assert.Error,
			want:      "source rejected: height too low: 350",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := MinimumHeight(tt.minHeight)(tt.stats)
			tt.wantErr(t, err)
			if err != nil {
				assert.Equal(t, tt.want, err.Error())
			}
		})
	}
}

func TestMinimumSourceBitrate(t *testing.T) {
	testCases := []struct {
		name    string
		stats   ffmpeg.VideoStats
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name:    "pass",
			stats:   testutil.MakeProbe("h264", 10000, 1080, time.Hour),
			wantErr: assert.NoError,
		},
		{
			name:    "unsupported codec",
			stats:   testutil.MakeProbe("unsupported", 10000, 1080, time.Hour),
			wantErr: assert.Error,
		},
		{
			name:    "bitrate too low",
			stats:   testutil.MakeProbe("h264", 1, 1080, time.Hour),
			wantErr: assert.Error,
		},
	}

	for _, tt := range testCases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			err := MinimumBitrate()(tt.stats)
			tt.wantErr(t, err)

			if err == nil {
				return
			}
			assert.ErrorIs(t, err, ErrSourceRejected{})
		})
	}
}
