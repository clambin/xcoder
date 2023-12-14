package ffmpeg

import (
	"errors"
	"strconv"
)

func getConvertArgsByOS(request Request) ([]string, error) {
	if request.VideoCodec != "hevc" {
		return nil, errors.New("only hvec supported right now")
	}
	request.VideoCodec = "hevc_videotoolbox"

	profile := "main"
	if request.BitsPerSample == 10 {
		profile = "main10"
	}

	return []string{
		"-y",
		"-nostats", "-loglevel", "error",
		//"-threads", "8",
		"-hwaccel", "videotoolbox",
		"-i", request.Source,
		"-map", "0",
		"-c:v", request.VideoCodec, "-profile:v", profile, "-b:v", strconv.Itoa(request.BitRate),
		"-c:a", "copy",
		"-c:s", "copy",
		"-f", "matroska",
	}, nil
}
