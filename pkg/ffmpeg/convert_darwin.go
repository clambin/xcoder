package ffmpeg

import (
	"errors"
	"strconv"
)

func makeConvertCommand(request Request, progressSocket string) (string, []string, error) {
	if request.VideoCodec != "hevc" {
		return "", nil, errors.New("only hvec supported right now")
	}
	request.VideoCodec = "hevc_videotoolbox"

	profile := "main"
	if request.BitsPerSample == 10 {
		profile = "main10"
	}

	ffmpegArgs := []string{
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
	}

	if progressSocket != "" {
		ffmpegArgs = append(ffmpegArgs,
			"-progress", "unix://"+progressSocket,
		)
	}

	ffmpegArgs = append(ffmpegArgs, request.Target)

	return "ffmpeg", ffmpegArgs, nil
}
