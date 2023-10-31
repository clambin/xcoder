package testutil

import (
	"github.com/clambin/vidconv/pkg/ffmpeg"
	"strconv"
	"time"
)

func MakeProbe(vcodec string, bitrate, height int, duration time.Duration) ffmpeg.Probe {
	return ffmpeg.Probe{
		Streams: []ffmpeg.Stream{{CodecType: "video", CodecName: vcodec, Height: height}},
		Format:  ffmpeg.Format{Duration: strconv.Itoa(int(duration.Seconds())), BitRate: strconv.Itoa(bitrate * 1024)},
	}
}
