package transcoder

import (
	"github.com/clambin/xcoder/ffmpeg"
)

// videoCodecs translates generic codec names into the OS-specific codec names
var videoCodecs = map[string]string{
	"hevc": "hevc_videotoolbox",
}

// inputArguments are added before the input chain.
// Use MacOS videotoolbox to decode video streams.
var inputArguments = ffmpeg.Args{
	"hwaccel": "videotoolbox",
}
