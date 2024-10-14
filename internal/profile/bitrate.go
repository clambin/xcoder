package profile

import (
	"fmt"
	"github.com/clambin/videoConvertor/internal/ffmpeg"
)

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

const lowQualityReduction = 0.8

func getMinimumBitRate(videoStats ffmpeg.VideoStats) (int, error) {
	rates, ok := minimumBitrates[videoStats.VideoCodec]
	if !ok {
		return 0, fmt.Errorf("invalid codec: %s", videoStats.VideoCodec)
	}
	return rates.getBitrate(videoStats.Height), nil
}
