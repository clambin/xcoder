package video

type minimumBitRate struct {
	height  int
	bitrate int
}

var minimumBitRates = map[string][]minimumBitRate{
	"hevc": {
		{height: 480, bitrate: 750 * 1024},
		{height: 720, bitrate: 1500 * 1024},
		{height: 1080, bitrate: 4000 * 1024},
		{height: 4000, bitrate: 15000 * 1024},
	},
}

func GetMinimumBitRate(file Video, codec string) (int, bool) {
	var bitRate int
	bitRates, ok := minimumBitRates[codec]
	if ok {
		for i := range bitRates {
			bitRate = bitRates[i].bitrate
			if file.Stats.Height() <= bitRates[i].height {
				break
			}
		}
	}
	return bitRate, ok
}
