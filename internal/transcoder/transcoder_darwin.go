package transcoder

import (
	"fmt"
	"strconv"

	"github.com/clambin/xcoder/ffmpeg"
)

func DecoderArguments(videoStats ffmpeg.VideoStats) []string {
	switch videoStats.VideoCodec {
	case "h264", "hevc":
		return []string{"-hwaccel", "videotoolbox"}
	default:
		return []string{}
	}
}

func encoderArguments(videoStats ffmpeg.VideoStats) ([]string, error) {
	switch videoStats.VideoCodec {
	case "hevc":
		profileName := "main"
		if videoStats.BitsPerSample == 10 {
			profileName = "main10"
		}
		return []string{
			"-c:v", "hevc_videotoolbox",
			"-b:v", strconv.Itoa(videoStats.BitRate),
			"-profile:v", profileName,
			"-c:a", "copy",
			"-c:s", "copy",
		}, nil
	default:
		return nil, fmt.Errorf("unsupported codec: %s", videoStats.VideoCodec)
	}
}
