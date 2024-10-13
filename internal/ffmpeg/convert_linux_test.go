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
			Source:      "foo.mkv",
			Target:      "foo.hevc.mkv",
			TargetStats: VideoStats{VideoCodec: "hevc", BitRate: 4_000_000, BitsPerSample: 8},
		},
		progressSocket: "socket",
		want:           "-i foo.mkv -b:v 4000000 -c:a copy -c:s copy -c:v libx265 -f matroska -profile:v main foo.hevc.mkv -nostats -loglevel error -progress unix://socket -y",
		wantErr:        assert.NoError,
	},
	{
		name: "hevc 10 bit",
		request: Request{
			Source:      "foo.mkv",
			Target:      "foo.hevc.mkv",
			TargetStats: VideoStats{VideoCodec: "hevc", BitRate: 4_000_000, BitsPerSample: 10},
		},
		progressSocket: "socket",
		want:           "-i foo.mkv -b:v 4000000 -c:a copy -c:s copy -c:v libx265 -f matroska -profile:v main10 foo.hevc.mkv -nostats -loglevel error -progress unix://socket -y",
		wantErr:        assert.NoError,
	},
	{
		name: "default is 8 bit",
		request: Request{
			Source:      "foo.mkv",
			Target:      "foo.hevc.mkv",
			TargetStats: VideoStats{VideoCodec: "hevc", BitRate: 4_000_000},
		},
		progressSocket: "socket",
		want:           "-i foo.mkv -b:v 4000000 -c:a copy -c:s copy -c:v libx265 -f matroska -profile:v main foo.hevc.mkv -nostats -loglevel error -progress unix://socket -y",
		wantErr:        assert.NoError,
	},
	{
		name: "only support for hevc",
		request: Request{
			Source:      "foo.mkv",
			Target:      "foo.hevc.mkv",
			TargetStats: VideoStats{VideoCodec: "h264", BitRate: 4_000_000, BitsPerSample: 8},
		},
		wantErr: assert.Error,
	},
}
