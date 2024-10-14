package profile

import (
	"fmt"
	"github.com/clambin/videoConvertor/internal/ffmpeg"
)

type Quality int

const (
	LowQuality Quality = iota
	HighQuality
	MaxQuality
)

var profiles = map[string]Profile{
	"hevc-low": {
		Codec:              "hevc",
		ConstantRateFactor: 28,
		Quality:            LowQuality,
		Rules: Rules{
			SkipCodec("hevc"),
			MinimumBitrate(LowQuality),
		},
	},
	"hevc-high": {
		Codec:              "hevc",
		ConstantRateFactor: 18,
		Quality:            HighQuality,
		Rules: Rules{
			SkipCodec("hevc"),
			MinimumBitrate(HighQuality),
		},
	},
	"hevc-max": {
		Codec:              "hevc",
		ConstantRateFactor: 10,
		Quality:            MaxQuality,
		Rules: Rules{
			SkipCodec("hevc"),
			MinimumHeight(720),
			MinimumBitrate(MaxQuality),
		},
	},
}

// A Profile serves two purposes. Firstly, it evaluates whether a source video file meets the requirements to be converted.
// Secondly, it determines the video parameters of the output video file.
type Profile struct {
	Codec              string
	Rules              Rules
	Quality            Quality
	Bitrate            int
	ConstantRateFactor int
}

// GetProfile returns the profile associated with name.
func GetProfile(name string) (Profile, error) {
	if profile, ok := profiles[name]; ok {
		return profile, nil
	}
	return Profile{}, fmt.Errorf("invalid profile name: %s", name)
}

// Evaluate verifies that the source's videoStats meet the profile's requirements. Otherwise it returns the first non-compliance.
func (p Profile) Evaluate(sourceVideoStats ffmpeg.VideoStats) error {
	return p.Rules.ShouldConvert(sourceVideoStats)
}
