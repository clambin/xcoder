package quality

import (
	"fmt"
	"github.com/clambin/vidconv/internal/server/requests"
	"github.com/clambin/vidconv/pkg/ffmpeg"
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
			MinimumSourceBitrate(),
		},
	},
	"hevc-high": {
		Codec:   "hevc",
		Quality: HighQuality,
		Rules: Rules{
			SkipCodec("hevc"),
			MinimumHeight(1080),
			MinimumSourceBitrate(),
		},
	},
	"hevc-max": {
		Codec:   "hevc",
		Quality: MaxQuality,
		Rules: Rules{
			SkipCodec("hevc"),
			MinimumHeight(1080),
			MinimumSourceBitrate(),
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
	if err := p.ShouldConvert(sourceStats); err != nil {
		return requests.Request{}, err
	}

	bitrate, ok := p.GetTargetBitrate(sourceStats)
	if !ok {
		// TODO: ShouldConvert should already have validated the source's codec
		panic("unable to get target bitrate from source stats")
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

func (p Profile) ShouldConvert(stats ffmpeg.VideoStats) error {
	return p.Rules.ShouldConvert(stats)
}

func (p Profile) GetTargetBitrate(source ffmpeg.VideoStats) (int, bool) {
	targetBitrate, ok := getMinimumBitrate(p.Codec, source.Height())
	switch p.Quality {
	case LowQuality:
	case HighQuality:
		var sourceMinimumBitrate int
		if sourceMinimumBitrate, ok = getMinimumBitrate(source.VideoCodec(), source.Height()); ok {
			oversized := float64(source.BitRate()) / float64(sourceMinimumBitrate)
			targetBitrate = int(float64(targetBitrate) * oversized)
		}
	case MaxQuality:
		// TODO: magic number
		targetBitrate = int(float64(source.BitRate()) / 1.2)
	}
	return targetBitrate, ok
}
