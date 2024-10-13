package ffmpeg

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestVideoStats_Read(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    VideoStats
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "valid",
			input: `{
    "streams": [
        { "codec_name": "hevc", "codec_type": "video", "bits_per_raw_sample": "10", "height": 1080, "width": 1920 },
        { "codec_type": "audio" },
        { "codec_type": "subtitle" }
    ],
    "format": { "filename": "foo.hevc.mkv", "duration": "1800.000", "bit_rate": "5000000" }
}
`,
			want:    VideoStats{Duration: 30 * time.Minute, VideoCodec: "hevc", BitRate: 5_000_000, BitsPerSample: 10, Height: 1080, Width: 1920},
			wantErr: assert.NoError,
		},
		{
			name:    "bitsPerSample defaults to 8",
			input:   `{"format": { "duration": "1800.000", "bit_rate": "5000000" }, "streams": [ { "codec_name": "hevc", "codec_type": "video", "height": 1080, "width": 1920 } ]}`,
			want:    VideoStats{Duration: 30 * time.Minute, VideoCodec: "hevc", BitRate: 5_000_000, BitsPerSample: 8, Height: 1080, Width: 1920},
			wantErr: assert.NoError,
		},
		{
			name:    "empty",
			wantErr: assert.Error,
		},
		{
			name:    "missing duration",
			input:   `{"format": {  }}`,
			wantErr: assert.Error,
		},
		{
			name:    "invalid duration",
			input:   `{"format": { "duration": "foobar" }}`,
			wantErr: assert.Error,
		},
		{
			name:    "missing bitrate",
			input:   `{"format": { "duration": "1800.00" }}`,
			wantErr: assert.Error,
		},
		{
			name:    "invalid bitrate",
			input:   `{"format": { "duration": "1800.000", "bit_rate": "foobar" }}`,
			wantErr: assert.Error,
		},
		{
			name:    "missing streams",
			input:   `{"format": { "duration": "1800.000", "bit_rate": "5000000" }}`,
			wantErr: assert.Error,
		},
		{
			name:    "missing video stream",
			input:   `{"format": { "duration": "1800.000", "bit_rate": "5000000" }, "streams": []}`,
			wantErr: assert.Error,
		},
		{
			name:    "invalid bitsPerSample defaults to 8",
			input:   `{"format": { "duration": "1800.000", "bit_rate": "5000000" }, "streams": [ { "codec_name": "hevc", "codec_type": "video", "bits_per_raw_sample": "foobar", "height": 1080, "width": 1920 } ]}`,
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := parse(tt.input)
			assert.Equal(t, tt.want, got)
			tt.wantErr(t, err)
		})
	}
}

func TestVideoStats_String(t *testing.T) {
	stats := VideoStats{
		Duration:      30 * time.Minute,
		VideoCodec:    "hevc",
		BitRate:       5_000_000,
		BitsPerSample: 10,
		Height:        1080,
	}
	const want = `hevc/1080/5.00 mbps`
	assert.Equal(t, want, stats.String())
	stats.VideoCodec = ""
	assert.Empty(t, stats.String())
}

// Current:
// BenchmarkParse-16         308307              3772 ns/op            1184 B/op         21 allocs/op
func BenchmarkParse(b *testing.B) {
	const input = `{
		"streams": [
	{ "codec_name": "hevc", "codec_type": "video", "bits_per_raw_sample": "10", "height": 1080, "width": 1920 },
	{ "codec_type": "audio" },
	{ "codec_type": "subtitle" }
	],
	"format": { "filename": "foo.hevc.mkv", "duration": "1800.000", "bit_rate": "5000000" }
}`

	for range b.N {
		if _, err := parse(input); err != nil {
			b.Fatal(err)
		}
	}
}
