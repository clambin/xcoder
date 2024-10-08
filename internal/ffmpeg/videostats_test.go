package ffmpeg

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"log/slog"
	"testing"
	"time"
)

func TestVideoStats(t *testing.T) {
	stats := VideoStats{
		Format: Format{
			BitRate:  "1024",
			Duration: "3600",
		},
		Streams: []Stream{
			{CodecType: "audio", CodecName: "aac"},
			{CodecType: "video", CodecName: "hevc", Height: 720, Width: 1280, BitsPerSample: 10},
		},
	}

	assert.Equal(t, 1024, stats.BitRate())
	assert.Equal(t, 10, stats.BitsPerSample())
	assert.Equal(t, time.Hour, stats.Duration())
	assert.Equal(t, "hevc", stats.VideoCodec())
	assert.Equal(t, 720, stats.Height())
	assert.Equal(t, 1280, stats.Width())
}

func TestNewVideoStats(t *testing.T) {
	stats := NewVideoStats("hevc", 1080, 3_000_000)
	assert.Equal(t, "hevc/1080/3.00 mbps", stats.String())
}

func TestVideoStats_LogValue(t *testing.T) {
	var buffer bytes.Buffer
	l := slog.New(slog.NewTextHandler(&buffer, &slog.HandlerOptions{ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
		if a.Key == slog.TimeKey {
			return slog.Attr{}
		}
		return a
	}}))

	stats := VideoStats{
		Format: Format{
			BitRate:  "1024000",
			Duration: "3600",
		},
		Streams: []Stream{
			{CodecType: "audio", CodecName: "aac"},
			{CodecType: "video", CodecName: "hevc", Height: 720, BitsPerSample: 10},
		},
	}

	l.Info("video", "stats", stats)
	assert.Equal(t, "level=INFO msg=video stats.codec=hevc stats.bitrate=1000 stats.depth=10 stats.height=720 stats.duration=1h0m0s\n", buffer.String())
}
