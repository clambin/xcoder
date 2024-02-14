package ffmpeg

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_makeConvertCommand(t *testing.T) {
	type args struct {
		request        Request
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
				request: Request{
					Source:        "foo.mkv",
					Target:        "foo.hevc.mkv",
					VideoCodec:    "hevc",
					BitsPerSample: 8,
					BitRate:       4000 * 1024,
				},
				progressSocket: "socket",
			},
			wantCmd:  "ffmpeg",
			wantArgs: []string{"-y", "-nostats", "-loglevel", "error", "-i", "foo.mkv", "-map", "0", "-c:v", "libx265", "-profile:v", "main", "-b:v", "4096000", "-c:a", "copy", "-c:s", "copy", "-f", "matroska", "-progress", "unix://socket", "foo.hevc.mkv"},
			wantErr:  assert.NoError,
		},
		{
			name: "hevc 10 bit",
			args: args{
				request: Request{
					Source:        "foo.mkv",
					Target:        "foo.hevc.mkv",
					VideoCodec:    "hevc",
					BitsPerSample: 10,
					BitRate:       4000 * 1024,
				},
				progressSocket: "socket",
			},
			wantCmd:  "ffmpeg",
			wantArgs: []string{"-y", "-nostats", "-loglevel", "error", "-i", "foo.mkv", "-map", "0", "-c:v", "libx265", "-profile:v", "main10", "-b:v", "4096000", "-c:a", "copy", "-c:s", "copy", "-f", "matroska", "-progress", "unix://socket", "foo.hevc.mkv"},
			wantErr:  assert.NoError,
		},
		{
			name: "default is 8 bit",
			args: args{
				request: Request{
					Source:        "foo.mkv",
					Target:        "foo.hevc.mkv",
					VideoCodec:    "hevc",
					BitsPerSample: 0,
					BitRate:       4000 * 1024,
				},
				progressSocket: "socket",
			},
			wantCmd:  "ffmpeg",
			wantArgs: []string{"-y", "-nostats", "-loglevel", "error", "-i", "foo.mkv", "-map", "0", "-c:v", "libx265", "-profile:v", "main", "-b:v", "4096000", "-c:a", "copy", "-c:s", "copy", "-f", "matroska", "-progress", "unix://socket", "foo.hevc.mkv"},
			wantErr:  assert.NoError,
		},
		{
			name: "only support for hevc",
			args: args{
				request: Request{
					Source:        "foo.mkv",
					Target:        "foo.hevc.mkv",
					VideoCodec:    "h264",
					BitsPerSample: 8,
					BitRate:       4000 * 1024,
				},
				progressSocket: "socket",
			},
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, got1, err := makeConvertCommand(tt.args.request, tt.args.progressSocket)
			tt.wantErr(t, err)
			if err != nil {
				return
			}
			assert.Equal(t, tt.wantCmd, got)
			assert.Equal(t, tt.wantArgs, got1)
		})
	}
}
