package cmd_test

import (
	"context"
	"github.com/clambin/videoConvertor/internal/ffmpeg/cmd"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFFMPEG_Build(t *testing.T) {
	command := cmd.Input("foo.mkv", cmd.Args{"hwaccel": "hevc_videotoolbox"}).
		Output("foo.hevc", cmd.Args{
			"c:v":       "hevc_videotoolbox",
			"profile:v": "main",
			"crf":       "10",
			"c:a":       "copy",
			"c:s":       "copy",
			"f":         "matroska",
		}).
		LogLevel("error").
		NoStats().
		OverWriteTarget(true).
		ProgressSocket("socket").
		AddGlobalArguments(cmd.Args{"foo": "bar"})

	want := `ffmeg -hwaccel hevc_videotoolbox -i foo.mkv -c:a copy -c:s copy -c:v hevc_videotoolbox -crf 10 -f matroska -profile:v main foo.hevc -foo bar -loglevel error -nostats -progress unix://socket -y`
	assert.Equal(t, want, command.Build(context.TODO()).String())
}
