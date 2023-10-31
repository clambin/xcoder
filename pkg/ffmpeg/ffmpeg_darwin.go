package ffmpeg

import (
	"errors"
	"strconv"
)

func makeConvertCommand(input, output, targetCodec string, bitrate int, progressSocket string) (string, []string, error) {
	if targetCodec != "hevc" {
		return "", nil, errors.New("only hvec supported right now")
	}
	targetCodec = "hevc_videotoolbox"

	ffmpegArgs := []string{
		"-y",
		//"-threads", "8",
		"-hwaccel", "videotoolbox",
		"-i", input,
		"-map", "0",
		"-c:v", targetCodec, "-profile:v", "main10", "-b:v", strconv.Itoa(bitrate),
		"-c:a", "copy",
		"-c:s", "copy",
	}

	if progressSocket != "" {
		ffmpegArgs = append(ffmpegArgs,
			"-progress", "unix://"+progressSocket,
		)
	}

	ffmpegArgs = append(ffmpegArgs, output)

	return "ffmpeg", ffmpegArgs, nil
}
