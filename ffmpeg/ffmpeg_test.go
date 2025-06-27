package ffmpeg_test

import (
	"strings"
	"testing"

	"github.com/clambin/xcoder/ffmpeg"
	"github.com/stretchr/testify/assert"
)

func TestFFMPEG_Build(t *testing.T) {
	tests := []struct {
		name string
		ff   *ffmpeg.FFMPEG
		want string
	}{
		{
			name: "simple",
			ff: ffmpeg.Decode("foo.mkv").
				Encode(
					"-c:v", "hevc_videotoolbox",
					"-b:v", "1000",
					"-profile:v", "main",
					"-c:a", "copy",
					"-c:s", "copy",
				).
				Muxer("matroska").
				Output("foo.hevc"),
			want: `-i foo.mkv -c:v hevc_videotoolbox -b:v 1000 -profile:v main -c:a copy -c:s copy -f matroska foo.hevc`,
		},
		{
			name: "no output",
			ff: ffmpeg.Decode("foo.mkv").
				Encode(
					"-c:v", "hevc_videotoolbox",
					"-profile:v", "main",
					"-c:a", "copy",
					"-c:s", "copy",
				).
				Muxer("matroska"),
			want: `-i foo.mkv -c:v hevc_videotoolbox -profile:v main -c:a copy -c:s copy -f matroska -`,
		},
		{
			name: "full example",
			ff: ffmpeg.Decode("foo.mkv", "-hwaccel", "videotoolbox").
				Encode(
					"-c:v", "hevc_videotoolbox",
					"-profile:v", "main",
					"-crf", "10",
					"-c:a", "copy",
					"-c:s", "copy",
				).
				Muxer("matroska").
				Output("foo.hevc").
				LogLevel("error").
				NoStats().
				OverWriteTarget().Progress(func(_ ffmpeg.Progress) {}, "/tmp/progress.sock"),
			want: `-hwaccel videotoolbox -i foo.mkv -c:v hevc_videotoolbox -profile:v main -crf 10 -c:a copy -c:s copy -f matroska -loglevel error -nostats -y -progress unix:///tmp/progress.sock foo.hevc`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, strings.Join(tt.ff.Build(t.Context()).Args[1:], " "))
		})
	}
}
