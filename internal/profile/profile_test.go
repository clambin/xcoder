package profile_test

import (
	"github.com/clambin/videoConvertor/internal/ffmpeg"
	"github.com/clambin/videoConvertor/internal/profile"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestProfile_Evaluate(t *testing.T) {
	tests := []struct {
		name           string
		profileName    string
		wantProfileErr assert.ErrorAssertionFunc
		sourceStats    ffmpeg.VideoStats
		wantEvalErr    assert.ErrorAssertionFunc
		targetStats    ffmpeg.VideoStats
	}{
		{
			name:           "invalid profile",
			profileName:    "invalid",
			wantProfileErr: assert.Error,
		},
		{
			name:           "source already in target codec",
			profileName:    "hevc-max",
			wantProfileErr: assert.NoError,
			sourceStats:    ffmpeg.NewVideoStats("hevc", 1080, 8_000_000),
			wantEvalErr:    assert.Error,
		},
		{
			name:           "source in unsupported codec",
			profileName:    "hevc-max",
			wantProfileErr: assert.NoError,
			sourceStats:    ffmpeg.NewVideoStats("invalid", 1080, 8_000_000),
			wantEvalErr:    assert.Error,
		},
		{
			name:           "height too low",
			profileName:    "hevc-max",
			wantProfileErr: assert.NoError,
			sourceStats:    ffmpeg.NewVideoStats("h264", 720, 8_000_000),
			wantEvalErr:    assert.Error,
		},
		{
			name:           "bitrate too low",
			profileName:    "hevc-max",
			wantProfileErr: assert.NoError,
			sourceStats:    ffmpeg.NewVideoStats("h264", 1080, 1_000_000),
			wantEvalErr:    assert.Error,
		},
		{
			name:           "success",
			profileName:    "hevc-max",
			wantProfileErr: assert.NoError,
			sourceStats:    ffmpeg.NewVideoStats("h264", 1080, 6_000_000),
			wantEvalErr:    assert.NoError,
			targetStats:    ffmpeg.NewVideoStats("hevc", 1080, 3_000_000),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			p, err := profile.GetProfile(tt.profileName)
			tt.wantProfileErr(t, err)
			if err != nil {
				return
			}
			stats, err := p.Evaluate(tt.sourceStats)
			tt.wantEvalErr(t, err)
			assert.Equal(t, tt.targetStats, stats)
		})
	}
}
