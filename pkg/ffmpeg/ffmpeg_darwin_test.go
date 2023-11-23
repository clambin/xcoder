package ffmpeg

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_makeConvertCommand(t *testing.T) {
	type args struct {
		input          string
		output         string
		targetCodec    string
		bitsPerSample  int
		bitrate        int
		progressSocket string
	}
	tests := []struct {
		name     string
		args     args
		wantCmd  string
		wantArgs []string
		wantErr  assert.ErrorAssertionFunc
	}{
		{
			name: "hevc 8 bit",
			args: args{
				input:          "foo.mkv",
				output:         "foo.hevc.mkv",
				targetCodec:    "hevc",
				bitsPerSample:  8,
				bitrate:        4000 * 1024,
				progressSocket: "socket",
			},
			wantCmd:  "ffmpeg",
			wantArgs: []string{"-y", "-hwaccel", "videotoolbox", "-i", "foo.mkv", "-map", "0", "-c:v", "hevc_videotoolbox", "-profile:v", "main", "-b:v", "4096000", "-c:a", "copy", "-c:s", "copy", "-progress", "unix://socket", "foo.hevc.mkv"},
			wantErr:  assert.NoError,
		},
		{
			name: "hevc 10 bit",
			args: args{
				input:          "foo.mkv",
				output:         "foo.hevc.mkv",
				targetCodec:    "hevc",
				bitsPerSample:  10,
				bitrate:        4000 * 1024,
				progressSocket: "socket",
			},
			wantCmd:  "ffmpeg",
			wantArgs: []string{"-y", "-hwaccel", "videotoolbox", "-i", "foo.mkv", "-map", "0", "-c:v", "hevc_videotoolbox", "-profile:v", "main10", "-b:v", "4096000", "-c:a", "copy", "-c:s", "copy", "-progress", "unix://socket", "foo.hevc.mkv"},
			wantErr:  assert.NoError,
		},
		{
			name: "default is 8 bit",
			args: args{
				input:          "foo.mkv",
				output:         "foo.hevc.mkv",
				targetCodec:    "hevc",
				bitsPerSample:  0,
				bitrate:        4000 * 1024,
				progressSocket: "socket",
			},
			wantCmd:  "ffmpeg",
			wantArgs: []string{"-y", "-hwaccel", "videotoolbox", "-i", "foo.mkv", "-map", "0", "-c:v", "hevc_videotoolbox", "-profile:v", "main", "-b:v", "4096000", "-c:a", "copy", "-c:s", "copy", "-progress", "unix://socket", "foo.hevc.mkv"},
			wantErr:  assert.NoError,
		},
		{
			name: "only support for hevc",
			args: args{
				input:          "foo.mkv",
				output:         "foo.out.mkv",
				targetCodec:    "h264",
				bitsPerSample:  8,
				bitrate:        4000 * 1024,
				progressSocket: "socket",
			},
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := makeConvertCommand(tt.args.input, tt.args.output, tt.args.targetCodec, tt.args.bitsPerSample, tt.args.bitrate, tt.args.progressSocket)
			tt.wantErr(t, err)
			if err != nil {
				return
			}
			assert.Equal(t, tt.wantCmd, got)
			assert.Equal(t, tt.wantArgs, got1)
		})
	}
}
