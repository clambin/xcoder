package ffmpeg

import (
	ffmpeg "github.com/u2takey/ffmpeg-go"
)

// videoCodecs translates generic codec names into the OS-specific codec names
var videoCodecs = map[string]string{
	"hevc": "libx265",
}

// inputArguments are added before the input chain.
var inputArguments = ffmpeg.KwArgs{}
