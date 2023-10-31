package rules

import (
	"github.com/clambin/vidconv/internal/video"
	"strconv"
)

type Rule func(video video.Video) (string, bool)

// SkipCodec rejects any video with the specified video codec
func SkipCodec(codec string) Rule {
	return func(video video.Video) (string, bool) {
		return "video already in " + codec, video.Stats.VideoCodec() != codec
	}
}

// OptimumBitrate rejects any video with a bitrate lower than the codec's optimum bitrate
func OptimumBitrate(codec string, margin float64) Rule {
	return func(input video.Video) (string, bool) {
		if bitRate, ok := video.GetMinimumBitRate(input, codec); ok {
			return "bitrate too low: " + strconv.Itoa(input.Stats.BitRate()/1024) + " kbps", input.Stats.BitRate() >= int(float64(bitRate)*margin)
		}
		return "unable to determine video bitrate", false
	}
}

// MinimumHeight rejects any video with a height lower than the specified height
func MinimumHeight(minHeight int) Rule {
	return func(input video.Video) (string, bool) {
		return "height too low: " + strconv.Itoa(input.Stats.Height()), input.Stats.Height() >= minHeight
	}
}
