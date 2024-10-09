package ffmpeg

import (
	"context"
	"github.com/stretchr/testify/assert"
	"log/slog"
	"strings"
	"testing"
)

func Test_makeConvertCommand(t *testing.T) {
	type args struct {
		request        Request
		progressSocket string
	}
	tests := []struct {
		name    string
		args    args
		wantCmd string
		wantErr assert.ErrorAssertionFunc
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
			wantCmd: "-i foo.mkv -b:v 4096000 -c:a copy -c:s copy -c:v libx265 -f matroska -profile:v main foo.hevc.mkv -nostats -loglevel error -progress unix://socket -y",
			wantErr: assert.NoError,
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
			wantCmd: "-i foo.mkv -b:v 4096000 -c:a copy -c:s copy -c:v libx265 -f matroska -profile:v main10 foo.hevc.mkv -nostats -loglevel error -progress unix://socket -y",
			wantErr: assert.NoError,
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
			wantCmd: "-i foo.mkv -b:v 4096000 -c:a copy -c:s copy -c:v libx265 -f matroska -profile:v main foo.hevc.mkv -nostats -loglevel error -progress unix://socket -y",
			wantErr: assert.NoError,
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
			},
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// ffmpeg-go.Silent() uses a global variable. :-(
			//t.Parallel()

			p := Processor{Logger: slog.Default()}
			ctx := context.WithValue(context.Background(), "test", "test")
			s, err := p.makeConvertCommand(ctx, tt.args.request, tt.args.progressSocket)
			tt.wantErr(t, err)
			if err != nil {
				return
			}
			clArgs := strings.Join(s.Compile().Args[1:], " ")
			assert.Equal(t, tt.wantCmd, clArgs)
			assert.Equal(t, "test", s.Context.Value("test"))
		})
	}
}
