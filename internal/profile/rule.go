package profile

import (
	"github.com/clambin/videoConvertor/internal/ffmpeg"
)

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
