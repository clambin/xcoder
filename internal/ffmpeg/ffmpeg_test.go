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
			request: Request{Source: "foo.mkv", Target: "foo.hevc.mkv", SourceStats: VideoStats{Height: 720, BitRate: 3_000_000, BitsPerSample: 8}, TargetVideoCodec: "hevc", ConstantRateFactor: 10},
			wantErr: assert.NoError,
		},
		{
			name:    "valid request - 10 bits per sample",
			request: Request{Source: "foo.mkv", Target: "foo.hevc.mkv", SourceStats: VideoStats{Height: 720, BitRate: 3_000_000, BitsPerSample: 10}, TargetVideoCodec: "hevc", ConstantRateFactor: 10},
			wantErr: assert.NoError,
		},
		{
			name:    "missing source",
			request: Request{Source: "", Target: "foo.hevc.mkv", SourceStats: VideoStats{Height: 720, BitRate: 3_000_000, BitsPerSample: 8}, TargetVideoCodec: "hevc", ConstantRateFactor: 10},
			wantErr: assert.Error,
		},
		{
			name:    "missing target",
			request: Request{Source: "foo.mkv", SourceStats: VideoStats{Height: 720, BitRate: 3_000_000, BitsPerSample: 8}, TargetVideoCodec: "hevc", ConstantRateFactor: 10},
			wantErr: assert.Error,
		},
		{
			name:    "missing codec",
			request: Request{Source: "foo.mkv", Target: "foo.hevc.mkv", SourceStats: VideoStats{Height: 720, BitRate: 3_000_000, BitsPerSample: 8}, ConstantRateFactor: 10},
			wantErr: assert.Error,
		},
		{
			name:    "missing height",
			request: Request{Source: "foo.mkv", Target: "foo.hevc.mkv", SourceStats: VideoStats{BitRate: 3_000_000, BitsPerSample: 8}, ConstantRateFactor: 10},
			wantErr: assert.Error,
		},
		{
			name:    "wrong codec",
			request: Request{Source: "foo.mkv", Target: "foo.hevc.mkv", SourceStats: VideoStats{Height: 720, BitRate: 3_000_000, BitsPerSample: 8}, TargetVideoCodec: "h264", ConstantRateFactor: 10},
			wantErr: assert.Error,
		},
		{
			name:    "wrong bits per sample",
			request: Request{Source: "foo.mkv", Target: "foo.hevc.mkv", SourceStats: VideoStats{Height: 720, BitRate: 3_000_000, BitsPerSample: 12}, TargetVideoCodec: "hevc", ConstantRateFactor: 10},
			wantErr: assert.Error,
		},
		{
			name:    "crf too low",
			request: Request{Source: "foo.mkv", Target: "foo.hevc.mkv", SourceStats: VideoStats{Height: 720, BitRate: 3_000_000}, TargetVideoCodec: "hevc", ConstantRateFactor: 0},
			wantErr: assert.Error,
		},
		{
			name:    "crf too high",
			request: Request{Source: "foo.mkv", Target: "foo.hevc.mkv", SourceStats: VideoStats{Height: 720, BitRate: 3_000_000, BitsPerSample: 12}, TargetVideoCodec: "hevc", ConstantRateFactor: 52},
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
