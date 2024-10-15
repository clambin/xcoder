package profile

import (
	"fmt"
	"github.com/clambin/videoConvertor/internal/ffmpeg"
)

var profiles = map[string]Profile{
	"hevc-low": {
		Codec:              "hevc",
		ConstantRateFactor: 28,
		Rules: Rules{
			SkipCodec("hevc"),
			MinimumBitrate(0.8),
		},
	},
	"hevc-medium": {
		Codec:              "hevc",
		ConstantRateFactor: 18,
		Rules: Rules{
			SkipCodec("hevc"),
			MinimumBitrate(1),
		},
	},
	"hevc-high": {
		Codec:              "hevc",
		ConstantRateFactor: 10,
		Rules: Rules{
			SkipCodec("hevc"),
			MinimumHeight(720),
			MinimumBitrate(1),
		},
	},
}

// A Profile serves two purposes. Firstly, it evaluates whether a source video file meets the requirements to be converted.
// Secondly, it determines the video parameters of the output video file.
type Profile struct {
	Codec              string
	Rules              Rules
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

////////////////////////////////////////////////////////////////////////////////////////////////////////////

type Rule func(stats ffmpeg.VideoStats) error

type Rules []Rule

func (r Rules) ShouldConvert(stats ffmpeg.VideoStats) error {
	for _, rule := range r {
		if err := rule(stats); err != nil {
			return err
		}
	}
	return nil
}

// SkipCodec rejects any video with the specified video codec
func SkipCodec(codec string) Rule {
	return func(stats ffmpeg.VideoStats) error {
		if stats.VideoCodec != codec {
			return nil
		}
		return ErrSourceInTargetCodec
	}
}

// MinimumBitrate rejects any source video with a bitrate lower than the codec's recommended bitrate.
// QualityFactor allows the profile to adjust the minimum bitrate (e.g. a low quality profile may permit a video at
// 80% of the minimum bitrate).
func MinimumBitrate(qualityFactor float64) Rule {
	return func(stats ffmpeg.VideoStats) error {
		minBitRate, err := getMinimumBitRate(stats)
		if err != nil {
			return ErrSourceRejected{Reason: err.Error()}
		}
		minBitRate = int(float64(minBitRate) * qualityFactor)
		if sourceBitRate := stats.BitRate; sourceBitRate < minBitRate {
			return ErrSourceRejected{Reason: "bitrate too low"}
		}
		return nil
	}
}

// MinimumHeight rejects any video with a height lower than the specified height
func MinimumHeight(minHeight int) Rule {
	return func(stats ffmpeg.VideoStats) error {
		if stats.Height < minHeight {
			return ErrSourceRejected{Reason: "height too low"}
		}
		return nil
	}
}
