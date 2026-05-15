package transcoder

import (
	"testing"

	"github.com/clambin/xcoder/ffmpeg"
	"github.com/stretchr/testify/assert"
)

func Test_buildTargetFilename(t *testing.T) {
	tests := []struct {
		name   string
		source File
		codec  string
		want   string
	}{
		{"movie", File{Path: "my-movie.mkv", VideoStats: ffmpeg.VideoStats{Height: 1080}}, "hevc", "my-movie.1080.hevc.mkv"},
		{"movie - no height", File{Path: "my-movie.mkv"}, "hevc", "my-movie.hevc.mkv"},
		{"movie - additional info", File{Path: "my movie 2026.foo.bar.mkv"}, "hevc", "my movie 2026.hevc.mkv"},
		{"movie - additional info - brackets", File{Path: "my movie (2026).foo.bar.mkv"}, "hevc", "my movie (2026).hevc.mkv"},
		{"episode", File{Path: "ep.s01e01.mkv", VideoStats: ffmpeg.VideoStats{Height: 720}}, "hevc", "ep.s01e01.720.hevc.mkv"},
		{"episode - no height", File{Path: "ep s01e01.mkv"}, "hevc", "ep.s01e01.hevc.mkv"},
		{"episode - additional info", File{Path: "ep.s01e01.1080p.BluRay.x264.foo.bar.mkv"}, "hevc", "ep.s01e01.hevc.mkv"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, buildTargetFilename(tt.source, tt.codec, "mkv"))
		})
	}
}
