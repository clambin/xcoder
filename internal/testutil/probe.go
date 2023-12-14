package testutil

import (
	"github.com/clambin/videoConvertor/pkg/ffmpeg"
	"strconv"
	"time"
)

func MakeProbe(codec string, bitrate, height int, duration time.Duration) ffmpeg.VideoStats {
	return ffmpeg.VideoStats{
		Streams: []ffmpeg.Stream{{CodecType: "video", CodecName: codec, Height: height}},
		Format:  ffmpeg.Format{Duration: strconv.Itoa(int(duration.Seconds())), BitRate: strconv.Itoa(bitrate * 1024)},
	}
}
