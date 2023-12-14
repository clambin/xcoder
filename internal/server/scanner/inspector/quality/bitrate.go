package quality

type minimumBitrate struct {
	height  int
	upper   int
	bitrate int
}

// https://www.yololiv.com/blog/h265-vs-h264-whats-the-difference-which-is-better/

var minimumBitrates = map[string][]minimumBitrate{
	"hevc": {
		{height: 480, upper: 480 + (720-480)/2, bitrate: 750 * 1024},
		{height: 720, upper: 720 + (1080-720)/2, bitrate: 1500 * 1024},
		{height: 1080, upper: 1080 + (2160 - 1080/2), bitrate: 3 * 1024 * 1024},
		{height: 2160, bitrate: 15 * 1024 * 1024},
	},
	"h264": {
		{height: 480, upper: 480 + (720-480)/2, bitrate: 1500 * 1024},
		{height: 720, upper: 720 + (1080-720)/2, bitrate: 3 * 1024 * 1024},
		{height: 1080, upper: 1080 + (2160-1080)/2, bitrate: 6 * 1024 * 1024},
		{height: 2160, bitrate: 32 * 1024 * 1024},
	},
}

func getMinimumBitrate(codec string, height int) (int, bool) {
	var bitRate int
	bitRates, ok := minimumBitrates[codec]
	if ok {
		for i := range bitRates {
			bitRate = bitRates[i].bitrate
			if height <= bitRates[i].upper {
				break
			}
		}
	}
	return bitRate, ok
}
