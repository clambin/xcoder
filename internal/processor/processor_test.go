package processor

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRequest_IsValid(t *testing.T) {
	tests := []struct {
		name    string
		request Request
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name:    "valid request - 8 bits per sample",
			request: Request{Source: "foo.mkv", Target: "foo.hevc.mkv", TargetStats: VideoStats{"hevc", 0, 3_000_000, 8, 0, 0}},
			wantErr: assert.NoError,
		},
		{
			name:    "valid request - 10 bits per sample",
			request: Request{Source: "foo.mkv", Target: "foo.hevc.mkv", TargetStats: VideoStats{"hevc", 0, 3_000_000, 10, 0, 0}},
			wantErr: assert.NoError,
		},
		{
			name:    "missing source",
			request: Request{Target: "foo.hevc.mkv", TargetStats: VideoStats{"hevc", 0, 3_000_000, 8, 0, 0}},
			wantErr: assert.Error,
		},
		{
			name:    "missing target",
			request: Request{Source: "foo.mkv", TargetStats: VideoStats{"hevc", 0, 3_000_000, 8, 0, 0}},
			wantErr: assert.Error,
		},
		{
			name:    "missing codec",
			request: Request{Source: "foo.mkv", Target: "foo.hevc.mkv", TargetStats: VideoStats{"", 0, 3_000_000, 8, 0, 0}},
			wantErr: assert.Error,
		},
		{
			name:    "wrong codec",
			request: Request{Source: "foo.mkv", Target: "foo.hevc.mkv", TargetStats: VideoStats{"h264", 0, 3_000_000, 8, 0, 0}},
			wantErr: assert.Error,
		},
		{
			name:    "wrong bits per sample",
			request: Request{Source: "foo.mkv", Target: "foo.hevc.mkv", TargetStats: VideoStats{"hevc", 0, 3_000_000, 16, 0, 0}},
			wantErr: assert.Error,
		},
		{
			name:    "missing bitrate",
			request: Request{Source: "foo.mkv", Target: "foo.hevc.mkv", TargetStats: VideoStats{"hevc", 0, 0, 8, 0, 0}},
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.wantErr(t, tt.request.IsValid())
		})
	}
}

func Test_makeConvertCommand(t *testing.T) {
	for _, tt := range makeConvertCommandTests {
		t.Run(tt.name, func(t *testing.T) {
			// ffmpeg-go.Silent() uses a global variable. :-(
			//t.Parallel()

			type ctxKey string
			key := ctxKey("test")
			ctx := context.WithValue(t.Context(), key, "test")

			s, err := makeConvertCommand(ctx, tt.request, tt.progressSocket)
			tt.wantErr(t, err)
			if err != nil {
				return
			}

			clArgs := strings.Join(s.Args[1:], " ")
			assert.Equal(t, tt.want, clArgs)
			// check that the command will be run with our context
			//assert.Equal(t, "test", s.Context.Value(key))
		})
	}
}
