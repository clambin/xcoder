package rules

import (
	"github.com/clambin/vidconv/internal/video"
)

type Rules []Rule

func MainProfile(codec string) Rules {
	return Rules{
		SkipCodec(codec),
		MinimumHeight(1080),
		OptimumBitrate(codec, 1.2),
	}
}

func (r Rules) ShouldConvert(video video.Video) (string, bool) {
	for _, rule := range r {
		if reason, ok := rule(video); !ok {
			return reason, false
		}
	}
	return "", true
}
