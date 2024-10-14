package ffmpeg

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRequest_IsValid(t *testing.T) {
	tests := []struct {
		name    string
		request Request
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name:    "valid request - 8 bits per sample",
			request: Request{nil, "foo.mkv", "foo.hevc.mkv", VideoStats{"hevc", 0, 3_000_000, 8, 0, 0}},
			wantErr: assert.NoError,
		},
		{
			name:    "valid request - 10 bits per sample",
			request: Request{nil, "foo.mkv", "foo.hevc.mkv", VideoStats{"hevc", 0, 3_000_000, 10, 0, 0}},
			wantErr: assert.NoError,
		},
		{
			name:    "missing source",
			request: Request{nil, "", "foo.hevc.mkv", VideoStats{"hevc", 0, 3_000_000, 8, 0, 0}},
			wantErr: assert.Error,
		},
		{
			name:    "missing target",
			request: Request{nil, "foo.mkv", "", VideoStats{"hevc", 0, 3_000_000, 8, 0, 0}},
			wantErr: assert.Error,
		},
		{
			name:    "missing codec",
			request: Request{nil, "foo.mkv", "foo.hevc.mkv", VideoStats{"", 0, 3_000_000, 8, 0, 0}},
			wantErr: assert.Error,
		},
		{
			name:    "wrong codec",
			request: Request{nil, "foo.mkv", "foo.hevc.mkv", VideoStats{"h264", 0, 3_000_000, 8, 0, 0}},
			wantErr: assert.Error,
		},
		{
			name:    "wrong bits per sample",
			request: Request{nil, "foo.mkv", "foo.hevc.mkv", VideoStats{"hevc", 0, 3_000_000, 16, 0, 0}},
			wantErr: assert.Error,
		},
		{
			name:    "missing bitrate",
			request: Request{nil, "foo.mkv", "foo.hevc.mkv", VideoStats{"hevc", 0, 0, 8, 0, 0}},
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.wantErr(t, tt.request.IsValid())
		})
	}
}
