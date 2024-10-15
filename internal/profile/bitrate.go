package profile

import (
	"fmt"
	"github.com/clambin/videoConvertor/internal/ffmpeg"
)

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

// getMinimumBitRate returns the recommended minimum bitrate given a video's codec and height.
func getMinimumBitRate(videoStats ffmpeg.VideoStats) (int, error) {
	rates, ok := minimumBitrates[videoStats.VideoCodec]
	if !ok {
		return 0, fmt.Errorf("invalid codec: %s", videoStats.VideoCodec)
	}
	return rates.getBitrate(videoStats.Height), nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////////

type bitRate struct {
	height  int
	bitrate int
}

type bitRates []bitRate

// getBitrate returns the minimum recommended bitrate given a video's height. If the height doesn't match any entries
// in minimumBitrate table, the bitrate is extrapolated between the lower and higher entries.
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
