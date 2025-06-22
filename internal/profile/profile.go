package profile

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/clambin/videoConvertor/ffmpeg"
)

var profiles = map[string]Profile{
	"hevc-low": {
		TargetCodec: "hevc",
		Rules: []Rule{
			SkipTargetCodec(),
		},
		CapBitrate: true,
	},
	"hevc-high": {
		TargetCodec: "hevc",
		Rules: []Rule{
			SkipTargetCodec(),
			SkipVideoHeight(1080),
		},
	},
}

type Profile struct {
	TargetCodec string
	Rules       []Rule
	CapBitrate  bool
}

// GetProfile returns the profile associated with name.
func GetProfile(name string) (Profile, error) {
	if profile, ok := profiles[name]; ok {
		return profile, nil
	}
	return Profile{}, fmt.Errorf("invalid profile name: %s", name)
}

func (p *Profile) Inspect(sourceVideoStats ffmpeg.VideoStats) (ffmpeg.VideoStats, error) {
	// evaluate all rules
	for _, rule := range p.Rules {
		if err := rule(p, sourceVideoStats); err != nil {
			return ffmpeg.VideoStats{}, err
		}
	}

	// evaluate minimum bitrate for the source. we check both source & target codec, as target codec may need
	// a higher bitrate than the source codec (e.g. hevc -> h264).
	minimumSourceBitrate, err := getMinimumBitRate(sourceVideoStats.VideoCodec, sourceVideoStats)
	if err != nil {
		return ffmpeg.VideoStats{}, &ErrSourceRejected{Reason: err.Error()}
	}
	minimumTargetBitrate, err := getMinimumBitRate(p.TargetCodec, sourceVideoStats)
	if err != nil {
		return ffmpeg.VideoStats{}, err
	}
	if minimumBitrate := max(minimumSourceBitrate, minimumTargetBitrate); sourceVideoStats.BitRate < minimumBitrate {
		return ffmpeg.VideoStats{}, &ErrSourceBitrateTooLow{Reason: "source bitrate must be at least " + strconv.Itoa(minimumBitrate) + " bps"}
	}

	// create target videoStats.  If we're asked to cap the bitrate, we ask for the minimum bitrate of the target codec.
	// otherwise, if the source bitrate is higher than the minimum, increate the target bitrate by the same factor.
	targetVideoStats := sourceVideoStats
	targetVideoStats.VideoCodec = p.TargetCodec
	targetVideoStats.BitRate = minimumTargetBitrate
	if !p.CapBitrate {
		oversampling := float64(sourceVideoStats.BitRate) / float64(minimumSourceBitrate)
		targetVideoStats.BitRate = int(float64(minimumTargetBitrate) * oversampling)
	}
	return targetVideoStats, nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type Rule func(profile *Profile, sourceStats ffmpeg.VideoStats) error

func SkipTargetCodec() Rule {
	return func(profile *Profile, sourceStats ffmpeg.VideoStats) error {
		if sourceStats.VideoCodec == profile.TargetCodec {
			return &ErrSourceRejected{Reason: "source video already in target codec"}
		}
		return nil
	}
}

func SkipVideoHeight(height int) Rule {
	return func(_ *Profile, sourceStats ffmpeg.VideoStats) error {
		if sourceStats.Height < height {
			return &ErrSourceRejected{Reason: "source video height is less than " + strconv.Itoa(height)}
		}
		return nil
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type ErrSourceRejected struct {
	Reason string
}

func (e *ErrSourceRejected) Error() string {
	return e.Reason
}

func (e *ErrSourceRejected) Is(e2 error) bool {
	var err *ErrSourceRejected
	ok := errors.As(e2, &err)
	return ok
}

type ErrSourceBitrateTooLow struct {
	Reason string
}

func (e *ErrSourceBitrateTooLow) Error() string {
	return e.Reason
}

func (e *ErrSourceBitrateTooLow) Is(e2 error) bool {
	var err *ErrSourceBitrateTooLow
	ok := errors.As(e2, &err)
	return ok
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type bitRate struct {
	height  int
	bitrate int
}

type bitRates []bitRate

func (b bitRates) getBitrate(height int) int {
	for i, r := range b {
		if r.height < height {
			continue
		}
		if r.height == height {
			return r.bitrate
		}
		if i == 0 {
			return r.bitrate
		}
		factor := float64(height-b[i-1].height) / float64(b[i].height-b[i-1].height)
		return b[i-1].bitrate + int(factor*(float64(b[i].bitrate-b[i-1].bitrate)))
	}
	return b[len(b)-1].bitrate
}

// https://www.yololiv.com/blog/h265-vs-h264-whats-the-difference-which-is-better/

var minimumBitrates = map[string]bitRates{
	"h264": {
		{height: 480, bitrate: 1500_000},
		{height: 720, bitrate: 3_000_000},
		{height: 1080, bitrate: 6_000_000},
		{height: 2160, bitrate: 32_000_000},
	},
	"hevc": {
		{height: 480, bitrate: 750_000},
		{height: 720, bitrate: 1500_000},
		{height: 1080, bitrate: 3_000_000},
		{height: 2160, bitrate: 15_000_000},
	},
}

func getMinimumBitRate(codec string, videoStats ffmpeg.VideoStats) (int, error) {
	rates, ok := minimumBitrates[codec]
	if !ok {
		return 0, fmt.Errorf("unsupported video codec: %s", codec)
	}
	rate := rates.getBitrate(videoStats.Height)
	return rate, nil
}
