package ffmpeg

import (
	"errors"
	"strconv"
)

func getConvertArgsByOS(request Request) ([]string, error) {
	// TODO: not yet tested on Linux
	if request.VideoCodec != "hevc" {
		return nil, errors.New("only hvec supported right now")
	}
	request.VideoCodec = "libx265"

	profile := "main"
	if request.BitsPerSample == 10 {
		profile = "main10"
	}

	return []string{
		"-y",
		"-nostats", "-loglevel", "error",
		//"-threads", "8",
		"-i", request.Source,
		"-map", "0",
		"-c:v", request.VideoCodec, "-profile:v", profile, "-b:v", strconv.Itoa(request.BitRate),
		"-c:a", "copy",
		"-c:s", "copy",
		"-f", "matroska",
	}, nil
}
