package profile

import (
	"github.com/clambin/videoConvertor/internal/convertor"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetProfile(t *testing.T) {
	for name := range profiles {
		_, err := GetProfile(name)
		assert.NoError(t, err)
	}

	_, err := GetProfile("invalid")
	assert.Error(t, err)
}

func TestProfile_Evaluate(t *testing.T) {
	tests := []struct {
		name        string
		sourceStats convertor.VideoStats
		wantEvalErr assert.ErrorAssertionFunc
		targetStats convertor.VideoStats
	}{
		{
			name:        "source already in target codec",
			sourceStats: convertor.VideoStats{VideoCodec: "hevc", Height: 1080, BitRate: 8_000_000},
			wantEvalErr: assert.Error,
		},
		{
			name:        "source in unsupported codec",
			sourceStats: convertor.VideoStats{VideoCodec: "invalid", Height: 1080, BitRate: 8_000_000},
			wantEvalErr: assert.Error,
		},
		{
			name:        "source in skipped codec",
			sourceStats: convertor.VideoStats{VideoCodec: "foobar", Height: 1080, BitRate: 8_000_000},
			wantEvalErr: assert.Error,
		},
		{
			name:        "height too low",
			sourceStats: convertor.VideoStats{VideoCodec: "h264", Height: 300, BitRate: 8_000_000},
			wantEvalErr: assert.Error,
		},
		{
			name:        "bitrate too low",
			sourceStats: convertor.VideoStats{VideoCodec: "h264", Height: 1080, BitRate: 1_000_000},
			wantEvalErr: assert.Error,
		},
		{
			name:        "minimum bitrate",
			sourceStats: convertor.VideoStats{VideoCodec: "h264", Height: 1080, BitRate: 6_000_000},
			wantEvalErr: assert.NoError,
			targetStats: convertor.VideoStats{VideoCodec: "hevc", Height: 1080, BitRate: 3_000_000},
		},
		{
			name:        "higher bitrate",
			sourceStats: convertor.VideoStats{VideoCodec: "h264", Height: 1080, BitRate: 12_000_000},
			wantEvalErr: assert.NoError,
			targetStats: convertor.VideoStats{VideoCodec: "hevc", Height: 1080, BitRate: 6_000_000},
		},
	}

	profile := Profile{
		Codec: "hevc",
		Rules: Rules{
			SkipCodec("foobar"),
			SkipTargetCodec(),
			MinimumHeight(1080),
			MinimumBitrate(),
		},
		Quality: MaxQuality,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			stats, err := profile.Evaluate(tt.sourceStats)
			tt.wantEvalErr(t, err)
			assert.Equal(t, tt.targetStats, stats)
		})
	}
}
