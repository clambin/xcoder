package ffmpeg

import (
	"log/slog"
)

// Processor implements video scanning (ffprobe) and converting (ffmpeg).  Implemented as a struct so that it can be
// mocked at the calling side.
type Processor struct {
	Logger *slog.Logger
}
