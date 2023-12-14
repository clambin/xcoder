package quality

import (
	"github.com/clambin/videoConvertor/pkg/ffmpeg"
	"strconv"
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
		if stats.VideoCodec() != codec {
			return nil
		}
		return ErrSourceRejected{Reason: "video already in target codec"}
	}
}

// MinimumSourceBitrate rejects any source video with a bitrate lower than the codec's recommended bitrate
func MinimumSourceBitrate() Rule {
	return func(stats ffmpeg.VideoStats) error {
		codec := stats.VideoCodec()
		bitRate, ok := getMinimumBitrate(codec, stats.Height())
		if !ok {
			return ErrSourceRejected{Reason: "unsupported codec: " + codec}
		}
		if stats.BitRate() < bitRate {
			return ErrSourceRejected{Reason: "bitrate too low: " + strconv.Itoa(stats.BitRate()/1024) + " kbps"}
		}
		return nil
	}
}

// MinimumHeight rejects any video with a height lower than the specified height
func MinimumHeight(minHeight int) Rule {
	return func(stats ffmpeg.VideoStats) error {
		if stats.Height() < minHeight {
			return ErrSourceRejected{Reason: "height too low: " + strconv.Itoa(stats.Height())}
		}
		return nil
	}
}
