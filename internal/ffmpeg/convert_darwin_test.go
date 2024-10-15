package ffmpeg

import (
	"github.com/stretchr/testify/assert"
)

var makeConvertCommandTests = []struct {
	name           string
	request        Request
	progressSocket string
	want           string
	wantErr        assert.ErrorAssertionFunc
}{
	{
		name: "hevc 8 bit",
		request: Request{
			Source:             "foo.mkv",
			Target:             "foo.hevc.mkv",
			SourceStats:        VideoStats{BitsPerSample: 8},
			TargetVideoCodec:   "hevc",
			ConstantRateFactor: 10,
		},
		progressSocket: "socket",
		want:           "-hwaccel videotoolbox -i foo.mkv -c:a copy -c:s copy -c:v hevc_videotoolbox -crf 10 -f matroska -profile:v main foo.hevc.mkv -loglevel error -nostats -progress unix://socket -y",
		wantErr:        assert.NoError,
	},
	{
		name: "hevc 10 bit",
		request: Request{
			Source:             "foo.mkv",
			Target:             "foo.hevc.mkv",
			SourceStats:        VideoStats{BitsPerSample: 10},
			TargetVideoCodec:   "hevc",
			ConstantRateFactor: 10,
		},
		progressSocket: "socket",
		want:           "-hwaccel videotoolbox -i foo.mkv -c:a copy -c:s copy -c:v hevc_videotoolbox -crf 10 -f matroska -profile:v main10 foo.hevc.mkv -loglevel error -nostats -progress unix://socket -y",
		wantErr:        assert.NoError,
	},
	{
		name: "default is 8 bit",
		request: Request{
			Source:             "foo.mkv",
			Target:             "foo.hevc.mkv",
			TargetVideoCodec:   "hevc",
			ConstantRateFactor: 10,
		},
		progressSocket: "socket",
		want:           "-hwaccel videotoolbox -i foo.mkv -c:a copy -c:s copy -c:v hevc_videotoolbox -crf 10 -f matroska -profile:v main foo.hevc.mkv -loglevel error -nostats -progress unix://socket -y",
		wantErr:        assert.NoError,
	},
	{
		name: "no progress socket",
		request: Request{
			Source:             "foo.mkv",
			Target:             "foo.hevc.mkv",
			TargetVideoCodec:   "hevc",
			ConstantRateFactor: 10,
		},
		want:    "-hwaccel videotoolbox -i foo.mkv -c:a copy -c:s copy -c:v hevc_videotoolbox -crf 10 -f matroska -profile:v main foo.hevc.mkv -loglevel error -nostats -y",
		wantErr: assert.NoError,
	},
	{
		name: "only support for hevc",
		request: Request{
			Source:             "foo.mkv",
			Target:             "foo.hevc.mkv",
			TargetVideoCodec:   "h264",
			ConstantRateFactor: 10,
		},
		wantErr: assert.Error,
	},
}
