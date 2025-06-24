package profile

import (
	"errors"
	"testing"

	"github.com/clambin/videoConvertor/ffmpeg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProfile_Inspect(t *testing.T) {
	tests := []struct {
		name            string
		profile         Profile
		source          ffmpeg.VideoStats
		wantErr         error
		wantTargetStats ffmpeg.VideoStats
	}{
		{
			name:    "wrong codec",
			profile: Profile{TargetCodec: "hevc", Rules: []Rule{SkipTargetCodec()}},
			source:  ffmpeg.VideoStats{VideoCodec: "hevc"},
			wantErr: &ErrSourceRejected{skip: true, reason: "source video already in target codec"},
		},
		{
			name:    "height too low",
			profile: Profile{TargetCodec: "hevc", Rules: []Rule{RejectVideoHeightTooLow(1080)}},
			source:  ffmpeg.VideoStats{VideoCodec: "hevc", Height: 800},
			wantErr: &ErrSourceRejected{reason: "source video height is less than 1080"},
		},
		{
			name:    "invalid source codec",
			profile: Profile{},
			source:  ffmpeg.VideoStats{VideoCodec: "invalid"},
			wantErr: errors.New("unsupported video codec: invalid"),
		},
		{
			name:    "invalid target codec",
			profile: Profile{TargetCodec: "invalid"},
			source:  ffmpeg.VideoStats{VideoCodec: "h264"},
			wantErr: errors.New("unsupported video codec: invalid"),
		},
		{
			name:    "source bitrate too low",
			profile: Profile{TargetCodec: "hevc", Rules: []Rule{RejectBitrateTooLow()}},
			source:  ffmpeg.VideoStats{VideoCodec: "h264", Height: 1080, BitRate: 4_000_000},
			wantErr: &ErrSourceRejected{reason: "source bitrate must be at least 6.00 mbps"},
		},
		{
			name:    "target bitrate too low",
			profile: Profile{TargetCodec: "h264", Rules: []Rule{RejectBitrateTooLow()}},
			source:  ffmpeg.VideoStats{VideoCodec: "hevc", Height: 1080, BitRate: 4_000_000},
			wantErr: &ErrSourceRejected{reason: "source bitrate must be at least 6.00 mbps"},
		},
		{
			name:            "valid source, capped",
			profile:         Profile{TargetCodec: "hevc", CapBitrate: true, Rules: []Rule{RejectBitrateTooLow(), SkipTargetCodec(), RejectVideoHeightTooLow(1080)}},
			source:          ffmpeg.VideoStats{VideoCodec: "h264", Height: 1080, BitRate: 8_000_000},
			wantTargetStats: ffmpeg.VideoStats{VideoCodec: "hevc", Height: 1080, BitRate: 3_000_000},
		},
		{
			name:            "valid source, not capped",
			profile:         Profile{TargetCodec: "hevc", Rules: []Rule{RejectBitrateTooLow(), SkipTargetCodec(), RejectVideoHeightTooLow(1080)}},
			source:          ffmpeg.VideoStats{VideoCodec: "h264", Height: 1080, BitRate: 8_000_000},
			wantTargetStats: ffmpeg.VideoStats{VideoCodec: "hevc", Height: 1080, BitRate: 4_000_000},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			targetVideoStats, err := tt.profile.Inspect(tt.source)
			if tt.wantErr != nil {
				require.Error(t, err)
				if errors.Is(tt.wantErr, &ErrSourceRejected{}) {
					assert.ErrorIs(t, err, tt.wantErr)
				}
				assert.Equal(t, tt.wantErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.wantTargetStats, targetVideoStats)
		})
	}
}

func Test_bitRates_getBitrate(t *testing.T) {
	b := bitRates{{height: 100, bitrate: 1000}, {height: 200, bitrate: 2000}, {height: 300, bitrate: 3000}}
	tests := []struct {
		name   string
		height int
		want   int
	}{
		{"too low", 50, 1000},
		{"match low", 200, 2000},
		{"interpolate", 250, 2500},
		{"match high", 300, 3000},
		{"too high", 400, 3000},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, b.getBitrate(tt.height))
		})
	}
}
