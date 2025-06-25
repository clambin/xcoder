package profile

import (
	"errors"
	"testing"

	"github.com/clambin/xcoder/ffmpeg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProfiles(t *testing.T) {
	valid1080 := ffmpeg.VideoStats{VideoCodec: "h264", BitRate: 8_000_000, Height: 1080}
	tooLow1080 := ffmpeg.VideoStats{VideoCodec: "h264", BitRate: 1_000_000, Height: 1080}
	valid720 := ffmpeg.VideoStats{VideoCodec: "h264", BitRate: 8_000_000, Height: 720}
	tooLow720 := ffmpeg.VideoStats{VideoCodec: "h264", BitRate: 750_000, Height: 720}
	valid360 := ffmpeg.VideoStats{VideoCodec: "h264", BitRate: 8_000_000, Height: 360}
	tooLow360 := ffmpeg.VideoStats{VideoCodec: "h264", BitRate: 100_000, Height: 360}

	tests := map[string][]struct {
		name    string
		source  ffmpeg.VideoStats
		wantErr assert.ErrorAssertionFunc
	}{
		"hevc-high": {
			{"valid 1080", valid1080, assert.NoError},
			{"1080 bitrate too low", tooLow1080, assert.Error},
			{"valid 720", valid720, assert.Error},
			{"720 bitrate too low", tooLow720, assert.Error},
			{"valid 360", valid360, assert.Error},
			{"360 bitrate too low", tooLow360, assert.Error},
		},
		"hevc-medium": {
			{"valid 1080", valid1080, assert.NoError},
			{"1080 bitrate too low", tooLow1080, assert.Error},
			{"valid 720", valid720, assert.NoError},
			{"720 bitrate too low", tooLow720, assert.Error},
			{"valid 360", valid360, assert.Error},
			{"360 bitrate too low", tooLow360, assert.Error},
		},
		"hevc-low": {
			{"valid 1080", valid1080, assert.NoError},
			{"1080 bitrate too low", tooLow1080, assert.Error},
			{"valid 720", valid720, assert.NoError},
			{"720 bitrate too low", tooLow720, assert.Error},
			{"valid 360", valid360, assert.NoError},
			{"360 bitrate too low", tooLow360, assert.Error},
		},
	}
	for name, cases := range tests {
		p, err := GetProfile(name)
		require.NoError(t, err)
		t.Run(name, func(t *testing.T) {
			for _, testCase := range cases {
				t.Run(testCase.name, func(t *testing.T) {
					_, err = p.Inspect(testCase.source)
					testCase.wantErr(t, err)
				})
			}
		})
	}
}

func TestGetProfile(t *testing.T) {
	tests := []struct {
		name      string
		wantErr   assert.ErrorAssertionFunc
		wantCodec string
	}{
		{"hevc-high", assert.NoError, "hevc"},
		{"hevc-medium", assert.NoError, "hevc"},
		{"hevc-low", assert.NoError, "hevc"},
		{"invalid", assert.Error, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profile, err := GetProfile(tt.name)
			tt.wantErr(t, err)
			assert.Equal(t, tt.wantCodec, profile.TargetCodec)
		})
	}
}

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
			wantErr: NewErrSourceRejected(true, "source video already in target codec"),
		},
		{
			name:    "height too low",
			profile: Profile{TargetCodec: "hevc", Rules: []Rule{RejectVideoHeightTooLow(1080)}},
			source:  ffmpeg.VideoStats{VideoCodec: "hevc", Height: 800},
			wantErr: NewErrSourceRejected(false, "source video height is less than 1080"),
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
			wantErr: NewErrSourceRejected(false, "source bitrate must be at least 6.0 mbps"),
		},
		{
			name:    "target bitrate too low",
			profile: Profile{TargetCodec: "h264", Rules: []Rule{RejectBitrateTooLow()}},
			source:  ffmpeg.VideoStats{VideoCodec: "hevc", Height: 1080, BitRate: 4_000_000},
			wantErr: NewErrSourceRejected(false, "source bitrate must be at least 6.0 mbps"),
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
