package profile

import (
	"errors"
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"

	"github.com/clambin/xcoder/ffmpeg"
)

var profiles = map[string]Profile{
	"hevc-low": {
		TargetCodec: "hevc",
		Rules: []Rule{
			SkipTargetCodec(),
			RejectBitrateTooLow(),
		},
		CapBitrate: true,
	},
	"hevc-medium": {
		TargetCodec: "hevc",
		Rules: []Rule{
			SkipTargetCodec(),
			RejectVideoHeightTooLow(720),
			RejectBitrateTooLow(),
		},
		CapBitrate: true,
	},
	"hevc-high": {
		TargetCodec: "hevc",
		Rules: []Rule{
			SkipTargetCodec(),
			RejectVideoHeightTooLow(1080),
			RejectBitrateTooLow(),
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
	return Profile{}, fmt.Errorf("invalid profile name: %q. supported profile names: %s", name, strings.Join(SupportedProfiles(), ", ")) //nolint:err113
}

func SupportedProfiles() []string {
	p := slices.Collect(maps.Keys(profiles))
	slices.Sort(p)
	return p
}

func (p *Profile) Inspect(sourceVideoStats ffmpeg.VideoStats) (ffmpeg.VideoStats, error) {
	// evaluate all rules
	for _, rule := range p.Rules {
		if err := rule(p, sourceVideoStats); err != nil {
			return ffmpeg.VideoStats{}, err
		}
	}

	// create target videoStats.  If we're asked to cap the bitrate, we ask for the minimum bitrate of the target codec.
	// otherwise, if the source bitrate is higher than the minimum, increate the target bitrate by the same factor.
	targetVideoStats := sourceVideoStats
	targetVideoStats.VideoCodec = p.TargetCodec
	var err error
	if targetVideoStats.BitRate, err = getTargetBitrate(sourceVideoStats, sourceVideoStats.VideoCodec, p.TargetCodec, p.CapBitrate); err != nil {
		return ffmpeg.VideoStats{}, err
	}
	return targetVideoStats, nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type Rule func(profile *Profile, sourceStats ffmpeg.VideoStats) error

func SkipTargetCodec() Rule {
	return func(profile *Profile, sourceStats ffmpeg.VideoStats) error {
		if sourceStats.VideoCodec == profile.TargetCodec {
			return &SourceRejectedError{skip: true, reason: "source video already in target codec"}
		}
		return nil
	}
}

func RejectVideoHeightTooLow(height int) Rule {
	return func(_ *Profile, sourceStats ffmpeg.VideoStats) error {
		if sourceStats.Height < height {
			return &SourceRejectedError{reason: "source video height is less than " + strconv.Itoa(height)}
		}
		return nil
	}
}

func RejectBitrateTooLow() Rule {
	return func(profile *Profile, sourceStats ffmpeg.VideoStats) error {
		// evaluate minimum bitrate for the source. we check both source & target codec, as target codec may need
		// a higher bitrate than the source codec (e.g. hevc -> h264).
		minimumBitrate, err := getMinimumBitRate(sourceStats, sourceStats.VideoCodec, profile.TargetCodec)
		if err != nil {
			return &SourceRejectedError{reason: err.Error()}
		}
		if sourceStats.BitRate < minimumBitrate {
			return &SourceRejectedError{reason: "source bitrate must be at least " + ffmpeg.Bits(minimumBitrate).Format(1)}
		}
		return nil
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type SourceRejectedError struct {
	reason string
	skip   bool
}

func NewErrSourceRejected(skip bool, reason string) *SourceRejectedError {
	return &SourceRejectedError{skip: skip, reason: reason}
}

func (e *SourceRejectedError) Skip() bool {
	return e.skip
}

func (e *SourceRejectedError) Error() string {
	return e.reason
}

func (e *SourceRejectedError) Is(e2 error) bool {
	var err *SourceRejectedError
	ok := errors.As(e2, &err)
	return ok && e.skip == err.skip && e.reason == err.reason
}

type UnsupportedCodecError struct {
	Codec string
}

func (e *UnsupportedCodecError) Error() string {
	return "unsupported codec: " + e.Codec
}

func (e *UnsupportedCodecError) Is(e2 error) bool {
	var err *UnsupportedCodecError
	ok := errors.As(e2, &err)
	return ok && e.Codec == err.Codec
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
		{height: 720, bitrate: 1_500_000},
		{height: 1080, bitrate: 3_000_000},
		{height: 2160, bitrate: 15_000_000},
	},
}

// getMinimumBitRate determines minimum bitrate for the source. we check both source & target codec, as target codec may need
// a higher bitrate than the source codec (e.g. hevc -> h264).
func getMinimumBitRate(videoStats ffmpeg.VideoStats, from string, to string) (int, error) {
	sourceMinimumBitrates, ok := minimumBitrates[from]
	if !ok {
		return 0, &UnsupportedCodecError{Codec: from}
	}
	targetMinimumBitrates, ok := minimumBitrates[to]
	if !ok {
		return 0, &UnsupportedCodecError{Codec: to}
	}
	return max(sourceMinimumBitrates.getBitrate(videoStats.Height), targetMinimumBitrates.getBitrate(videoStats.Height)), nil
}

func getTargetBitrate(videoStats ffmpeg.VideoStats, from string, to string, capBitrate bool) (int, error) {
	sourceMinimumBitrates, ok := minimumBitrates[from]
	if !ok {
		return 0, &UnsupportedCodecError{Codec: from}
	}
	targetMinimumBitrates, ok := minimumBitrates[to]
	if !ok {
		return 0, &UnsupportedCodecError{Codec: to}
	}
	bitrate := targetMinimumBitrates.getBitrate(videoStats.Height)
	if !capBitrate {
		oversampling := float64(videoStats.BitRate) / float64(sourceMinimumBitrates.getBitrate(videoStats.Height))
		bitrate = int(float64(bitrate) * oversampling)
	}
	return bitrate, nil
}
