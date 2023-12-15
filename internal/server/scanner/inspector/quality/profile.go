package quality

import (
	"errors"
	"fmt"
	"github.com/clambin/videoConvertor/internal/server/requests"
	"github.com/clambin/videoConvertor/pkg/ffmpeg"
)

type Quality int

const (
	LowQuality Quality = iota
	HighQuality
	MaxQuality
)

var profiles = map[string]Profile{
	"hevc-low": {
		Codec:   "hevc",
		Quality: LowQuality,
		Rules: Rules{
			SkipCodec("hevc"),
			MinimumBitrate(),
		},
	},
	"hevc-high": {
		Codec:   "hevc",
		Quality: HighQuality,
		Rules: Rules{
			SkipCodec("hevc"),
			MinimumHeight(1080),
			MinimumBitrate(),
		},
	},
	"hevc-max": {
		Codec:   "hevc",
		Quality: MaxQuality,
		Rules: Rules{
			SkipCodec("hevc"),
			MinimumHeight(1080),
			MinimumBitrate(),
		},
	},
}

type Profile struct {
	Codec   string
	Quality Quality
	Bitrate int
	Rules   Rules
}

func GetProfile(name string) (Profile, error) {
	if profile, ok := profiles[name]; ok {
		return profile, nil
	}
	return Profile{}, fmt.Errorf("invalid profile name: %s", name)
}

func (p Profile) MakeRequest(target, source string, sourceStats ffmpeg.VideoStats) (requests.Request, error) {
	if err := p.Rules.ShouldConvert(sourceStats); err != nil {
		return requests.Request{}, err
	}

	bitrate, ok := p.GetTargetBitrate(sourceStats)
	if !ok {
		// ShouldConvert will have already validated the source's codec, so this should never happen.
		return requests.Request{}, errors.New("unable to get target bitrate from source stats")
	}

	return requests.Request{
		Request: ffmpeg.Request{
			Source:        source,
			Target:        target,
			VideoCodec:    p.Codec,
			BitsPerSample: sourceStats.BitsPerSample(),
			BitRate:       bitrate,
		},
		SourceStats: sourceStats,
	}, nil
}

func (p Profile) GetTargetBitrate(source ffmpeg.VideoStats) (int, bool) {
	targetBitrate, ok := GetMinimumBitrate(p.Codec, source.Height())
	switch p.Quality {
	case LowQuality:
	case HighQuality:
		// for high quality, we oversize the target's minimum bitrate by the ratio between the source's minimum & actual bitrates
		// e.g. if the target's minimum bitrate is 100, the source's minimum bitrate is 200 and the actual bitrate is 400,
		// then the target bitrate is 100 * ( 400 / 200) = 200
		var sourceMinimumBitrate int
		if sourceMinimumBitrate, ok = GetMinimumBitrate(source.VideoCodec(), source.Height()); ok {
			oversized := float64(source.BitRate()) / float64(sourceMinimumBitrate)
			targetBitrate = int(float64(targetBitrate) * oversized)
		}
	case MaxQuality:
		// for max quality, set the bitrate so that it does not cause an increase in file size between source & target video.
		// Note: this value is calculated using a couple of sample videos and most definitely "wrong".
		targetBitrate = int(float64(source.BitRate()) / 1.2)
	}
	return targetBitrate, ok
}
