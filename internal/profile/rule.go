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

// MinimumBitrate rejects any source video with a bitrate lower than the codec's recommended bitrate for the provided Quality
func MinimumBitrate(quality Quality) Rule {
	return func(stats ffmpeg.VideoStats) error {
		minBitRate, err := getMinimumBitRate(stats)
		if err != nil {
			return ErrSourceRejected{Reason: err.Error()}
		}
		if quality == LowQuality {
			minBitRate = int(float64(minBitRate) * 0.8)
		}
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
