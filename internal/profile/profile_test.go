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
			sourceStats:    ffmpeg.VideoStats{VideoCodec: "hevc", Height: 1080, BitRate: 8_000_000},
			wantEvalErr:    assert.Error,
		},
		{
			name:           "source in unsupported codec",
			profileName:    "hevc-max",
			wantProfileErr: assert.NoError,
			sourceStats:    ffmpeg.VideoStats{VideoCodec: "invalid", Height: 1080, BitRate: 8_000_000},
			wantEvalErr:    assert.Error,
		},
		{
			name:           "height too low",
			profileName:    "hevc-max",
			wantProfileErr: assert.NoError,
			sourceStats:    ffmpeg.VideoStats{VideoCodec: "h264", Height: 300, BitRate: 8_000_000},
			wantEvalErr:    assert.Error,
		},
		{
			name:           "bitrate too low",
			profileName:    "hevc-low",
			wantProfileErr: assert.NoError,
			sourceStats:    ffmpeg.VideoStats{VideoCodec: "h264", Height: 1080, BitRate: 1_000_000},
			wantEvalErr:    assert.Error,
		},
		{
			name:           "success",
			profileName:    "hevc-max",
			wantProfileErr: assert.NoError,
			sourceStats:    ffmpeg.VideoStats{VideoCodec: "h264", Height: 1080, BitRate: 6_000_000},
			wantEvalErr:    assert.NoError,
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
			tt.wantEvalErr(t, p.Evaluate(tt.sourceStats))
		})
	}
}
