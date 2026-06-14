package transcoder

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
		//CapBitrate: true,
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

// SourceRejectedError is returned when a source media file is rejected by the profile
type SourceRejectedError struct {
	Reason string
}

func (e *SourceRejectedError) Error() string {
	return e.Reason
}

func (e *SourceRejectedError) Is(e2 error) bool {
	var err *SourceRejectedError
	ok := errors.As(e2, &err)
	return ok && e.Reason == err.Reason
}

// SourceSkippedError is returned when a source media file is skipped by the profile
type SourceSkippedError struct {
	Reason string
}

func (e *SourceSkippedError) Error() string {
	return e.Reason
}

func (e *SourceSkippedError) Is(e2 error) bool {
	var err *SourceSkippedError
	ok := errors.As(e2, &err)
	return ok && e.Reason == err.Reason
}

// A Profile specifies the requirements of a source media file and the corresponding converted target media file.
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

// SupportedProfiles returns a sorted list of supported profile names
func SupportedProfiles() []string {
	p := slices.Collect(maps.Keys(profiles))
	slices.Sort(p)
	return p
}

// Analyze performs a profile analysis on the source file and returns the target video stats
func (p Profile) Analyze(source File) (ffmpeg.VideoStats, error) {
	// evaluate all rules
	for _, rule := range p.Rules {
		if err := rule(p, source.VideoStats); err != nil {
			return ffmpeg.VideoStats{}, err
		}
	}

	// determine target videoStats
	targetVideoStats := source.VideoStats
	targetVideoStats.VideoCodec = p.TargetCodec
	var err error
	if targetVideoStats.BitRate, err = getTargetBitrate(source.VideoStats, source.VideoStats.VideoCodec, p.TargetCodec, p.CapBitrate); err != nil {
		return ffmpeg.VideoStats{}, err
	}
	return targetVideoStats, nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// A Rule is a function that evaluates a source file and returns an error if the source file does not meet the profile requirements
type Rule func(profile Profile, sourceStats ffmpeg.VideoStats) error

// SkipTargetCodec returns a rule that skips the profile if the source video is already in the target codec
func SkipTargetCodec() Rule {
	return func(profile Profile, sourceStats ffmpeg.VideoStats) error {
		if sourceStats.VideoCodec == profile.TargetCodec {
			return &SourceSkippedError{Reason: "source video already in target codec"}
		}
		return nil
	}
}

// RejectVideoHeightTooLow returns a rule that rejects the profile if the source video height is too low
func RejectVideoHeightTooLow(height int) Rule {
	return func(_ Profile, sourceStats ffmpeg.VideoStats) error {
		if sourceStats.Height < height {
			return &SourceRejectedError{Reason: "source video height is less than " + strconv.Itoa(height)}
		}
		return nil
	}
}

// RejectBitrateTooLow returns a rule that rejects the profile if the source video bitrate is too low
func RejectBitrateTooLow() Rule {
	return func(profile Profile, sourceStats ffmpeg.VideoStats) error {
		// Determine the minimum bitrate for the video codecs.
		minimumBitrate, err := getMinimumBitRate(sourceStats, sourceStats.VideoCodec, profile.TargetCodec)
		if err != nil {
			return &SourceRejectedError{Reason: err.Error()}
		}
		// Return an error is the source has a lower bitrate.
		if sourceStats.BitRate < minimumBitrate {
			return &SourceRejectedError{Reason: "source bitrate must be at least " + ffmpeg.Bits(minimumBitrate).Format(1)}
		}
		return nil
	}
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

// getMinimumBitRate determines the minimum bitrate. we check both source and target codec, as target codec may need
// a higher bitrate than the source codec (e.g. hevc -> h264).
func getMinimumBitRate(videoStats ffmpeg.VideoStats, from string, to string) (int, error) {
	sourceMinimumBitrates, ok := minimumBitrates[from]
	if !ok {
		return 0, &SourceRejectedError{Reason: "unsupported source video codec: " + from}
	}
	targetMinimumBitrates, ok := minimumBitrates[to]
	if !ok {
		return 0, &SourceRejectedError{Reason: "unsupported target video codec: " + to}
	}
	return max(sourceMinimumBitrates.getBitrate(videoStats.Height), targetMinimumBitrates.getBitrate(videoStats.Height)), nil
}

func getTargetBitrate(videoStats ffmpeg.VideoStats, from string, to string, capBitrate bool) (int, error) {
	// minimum bitrate for the source codec
	sourceMinimumBitrates, ok := minimumBitrates[from]
	if !ok {
		return 0, &SourceRejectedError{Reason: "unsupported source video codec: " + from}
	}
	// minimum bitrates for the target codec
	targetMinimumBitrates, ok := minimumBitrates[to]
	if !ok {
		return 0, &SourceRejectedError{Reason: "unsupported target video codec: " + to}
	}
	// minimum bitrate for the video's height
	bitrate := targetMinimumBitrates.getBitrate(videoStats.Height)
	if capBitrate {
		return bitrate, nil
	}
	// if we're not capping the bitRate at the minimum, determine the oversampling factor,
	// i.e., by how much the source is over the minimum rate for the source code & height
	oversampling := float64(videoStats.BitRate) / float64(sourceMinimumBitrates.getBitrate(videoStats.Height))
	// apply the oversampling factor to the target codec's minimum bitrate
	// so, if the source is twice its minimum bitrate, the target will also be twice its minimum bitrate
	return int(float64(bitrate) * oversampling), nil
}
