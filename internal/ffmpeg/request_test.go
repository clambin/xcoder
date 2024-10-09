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
			request: Request{"foo.mkv", "foo.hevc.mkv", "hevc", 8, 3_000_000, nil},
			wantErr: assert.NoError,
		},
		{
			name:    "valid request - 10 bits per sample",
			request: Request{"foo.mkv", "foo.hevc.mkv", "hevc", 10, 3_000_000, nil},
			wantErr: assert.NoError,
		},
		{
			name:    "missing source",
			request: Request{"", "foo.hevc.mkv", "hevc", 8, 3_000_000, nil},
			wantErr: assert.Error,
		},
		{
			name:    "missing target",
			request: Request{"foo.mkv", "", "hevc", 8, 3_000_000, nil},
			wantErr: assert.Error,
		},
		{
			name:    "missing codec",
			request: Request{"foo.mkv", "foo.hevc.mkv", "", 8, 3_000_000, nil},
			wantErr: assert.Error,
		},
		{
			name:    "wrong codec",
			request: Request{"foo.mkv", "foo.hevc.mkv", "h264", 8, 3_000_000, nil},
			wantErr: assert.Error,
		},
		{
			name:    "wrong bits per sample",
			request: Request{"foo.mkv", "foo.hevc.mkv", "hevc", 4, 3_000_000, nil},
			wantErr: assert.Error,
		},
		{
			name:    "missing bitrate",
			request: Request{"foo.mkv", "foo.hevc.mkv", "hevc", 8, 0, nil},
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
