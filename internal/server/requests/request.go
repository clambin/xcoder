package requests

import (
	"github.com/clambin/vidconv/pkg/ffmpeg"
	"log/slog"
)

/*
	type Request struct {
		Source        string
		SourceStats   ffmpeg.VideoStats
		Target        string
		Codec         string
		BitRate       int
		BitsPerSample int
	}
*/
type Request struct {
	ffmpeg.Request
	SourceStats ffmpeg.VideoStats
}

func (r Request) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Group("source",
			slog.String("filename", r.Source),
			slog.String("codec", r.SourceStats.VideoCodec()),
			slog.Int("bitrate", r.SourceStats.BitRate()),
			slog.Int("bitsPerSample", r.SourceStats.BitsPerSample()),
		),
		slog.Group("target",
			slog.String("filename", r.Target),
			slog.String("codec", r.VideoCodec),
			slog.Int("bitrate", r.BitRate),
			slog.Int("bitsPerSample", r.BitsPerSample),
		),
	)
}
