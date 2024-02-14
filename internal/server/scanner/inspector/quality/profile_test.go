package quality_test

import (
	"github.com/clambin/videoConvertor/internal/server/scanner/inspector/quality"
	"github.com/clambin/videoConvertor/internal/testutil"
	"github.com/clambin/videoConvertor/pkg/ffmpeg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestProfile_MakeRequest(t *testing.T) {
	tests := []struct {
		name        string
		profile     string
		sourceStats ffmpeg.VideoStats
		wantErr     assert.ErrorAssertionFunc
		wantBitrate int
	}{
		{
			name:        "low - pass",
			profile:     "hevc-low",
			sourceStats: testutil.MakeProbe("h264", 3*1024, 720, time.Hour),
			wantErr:     assert.NoError,
			wantBitrate: 1500 * 1024,
		},
		{
			name:        "low - bitrate too low",
			profile:     "hevc-low",
			sourceStats: testutil.MakeProbe("h264", 2*1024, 720, time.Hour),
			wantErr:     assert.Error,
		},
		{
			name:        "low - already hevc",
			profile:     "hevc-low",
			sourceStats: testutil.MakeProbe("hevc", 2*1024, 720, time.Hour),
			wantErr:     assert.Error,
		},
		{
			name:        "high - pass",
			profile:     "hevc-high",
			sourceStats: testutil.MakeProbe("h264", 6*1024, 1080, time.Hour),
			wantErr:     assert.NoError,
			wantBitrate: 3 * 1024 * 1024,
		},
		{
			name:        "high - oversize",
			profile:     "hevc-high",
			sourceStats: testutil.MakeProbe("h264", 12*1024, 1080, time.Hour),
			wantErr:     assert.NoError,
			wantBitrate: 6 * 1024 * 1024,
		},
		{
			name:        "high - bitrate too low",
			profile:     "hevc-high",
			sourceStats: testutil.MakeProbe("h264", 2*1024, 1080, time.Hour),
			wantErr:     assert.Error,
		},
		{
			name:        "high - height too low",
			profile:     "hevc-high",
			sourceStats: testutil.MakeProbe("h264", 12*1024, 720, time.Hour),
			wantErr:     assert.Error,
		},
		{
			name:        "high - already hevc",
			profile:     "hevc-high",
			sourceStats: testutil.MakeProbe("hevc", 6*1024, 1080, time.Hour),
			wantErr:     assert.Error,
		},
		{
			name:        "max - pass",
			profile:     "hevc-max",
			sourceStats: testutil.MakeProbe("h264", 6*1024, 1080, time.Hour),
			wantErr:     assert.NoError,
			wantBitrate: 5 * 1024 * 1024,
		},
		{
			name:        "max - oversize",
			profile:     "hevc-max",
			sourceStats: testutil.MakeProbe("h264", 12*1024, 1080, time.Hour),
			wantErr:     assert.NoError,
			wantBitrate: 10 * 1024 * 1024,
		},
		{
			name:        "max - bitrate too low",
			profile:     "hevc-max",
			sourceStats: testutil.MakeProbe("h264", 2*1024, 1080, time.Hour),
			wantErr:     assert.Error,
		},
		{
			name:        "max - height too low",
			profile:     "hevc-max",
			sourceStats: testutil.MakeProbe("h264", 12*1024, 720, time.Hour),
			wantErr:     assert.Error,
		},
		{
			name:        "max - already hevc",
			profile:     "hevc-max",
			sourceStats: testutil.MakeProbe("hevc", 6*1024, 1080, time.Hour),
			wantErr:     assert.Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			p, err := quality.GetProfile(tt.profile)
			require.NoError(t, err)

			req, err := p.MakeRequest("target.mkv", "source.mkv", tt.sourceStats)
			tt.wantErr(t, err)
			if err != nil {
				t.Log(err)
				return
			}
			assert.Equal(t, tt.sourceStats.BitsPerSample(), req.BitsPerSample)
			assert.Equal(t, tt.wantBitrate, req.BitRate)
		})
	}
}
