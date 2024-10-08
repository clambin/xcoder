package ffmpeg

// videoCodecs translates generic codec names into the OS-specific codec names
var videoCodecs = map[string]string{
	"hevc": "hevc_videotoolbox",
}

// prefixArguments are added before the input chain.
// Use MacOS videotoolbox to decode video streams.
var prefixArguments = []string{
	"-hwaccel", "videotoolbox",
}
