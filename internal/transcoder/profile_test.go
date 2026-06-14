package transcoder

import (
	"testing"

	"github.com/clambin/xcoder/ffmpeg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProfile_Analyze(t *testing.T) {
	tests := []struct {
		name    string
		profile string
		source  File
		want    ffmpeg.VideoStats
		err     error
	}{
		{"not high enough", "hevc-high", File{VideoStats: ffmpeg.VideoStats{VideoCodec: "h264"}}, ffmpeg.VideoStats{}, &SourceRejectedError{Reason: "source video height is less than 1080"}},
		{"bitrate too low", "hevc-high", File{VideoStats: ffmpeg.VideoStats{VideoCodec: "h264", Height: 1080}}, ffmpeg.VideoStats{}, &SourceRejectedError{Reason: "source bitrate must be at least 6.0 mbps"}},
		{"unsupported codec", "hevc-high", File{VideoStats: ffmpeg.VideoStats{VideoCodec: "invalid", Height: 1080}}, ffmpeg.VideoStats{}, &SourceRejectedError{Reason: "unsupported source video codec: invalid"}},
		{"already in target codec", "hevc-high", File{VideoStats: ffmpeg.VideoStats{VideoCodec: "hevc"}}, ffmpeg.VideoStats{}, &SourceSkippedError{Reason: "source video already in target codec"}},
		{"minimum bitrate", "hevc-high", File{VideoStats: ffmpeg.VideoStats{VideoCodec: "h264", Height: 1080, BitRate: 8_000_000}}, ffmpeg.VideoStats{VideoCodec: "hevc", Height: 1080, BitRate: 4_000_000}, nil},
		{"oversampled", "hevc-high", File{VideoStats: ffmpeg.VideoStats{VideoCodec: "h264", Height: 1080, BitRate: 16_000_000}}, ffmpeg.VideoStats{VideoCodec: "hevc", Height: 1080, BitRate: 8_000_000}, nil},
		{"hevc-medium", "hevc-medium", File{VideoStats: ffmpeg.VideoStats{VideoCodec: "h264", Height: 872, BitRate: 10_080_000}}, ffmpeg.VideoStats{VideoCodec: "hevc", Height: 872, BitRate: 5_040_000}, nil},
		{"hevc-low", "hevc-low", File{VideoStats: ffmpeg.VideoStats{VideoCodec: "h264", Height: 872, BitRate: 10_080_000}}, ffmpeg.VideoStats{VideoCodec: "hevc", Height: 872, BitRate: 2_133_333}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := GetProfile(tt.profile)
			require.NoError(t, err)
			got, err := p.Analyze(tt.source)
			if tt.err != nil {
				assert.ErrorIs(t, err, tt.err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
