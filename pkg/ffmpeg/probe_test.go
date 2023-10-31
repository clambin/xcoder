package ffmpeg

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"log/slog"
	"testing"
	"time"
)

func TestProbe(t *testing.T) {
	probe := Probe{
		Format: Format{
			BitRate:  "1024",
			Duration: "3600",
		},
		Streams: []Stream{
			{CodecType: "audio", CodecName: "aac"},
			{CodecType: "video", CodecName: "hevc", Height: 720},
		},
	}

	assert.Equal(t, 1024, probe.BitRate())
	assert.Equal(t, time.Hour, probe.Duration())
	assert.Equal(t, "hevc", probe.VideoCodec())
	assert.Equal(t, 720, probe.Height())
}

func TestProbe_LogValue(t *testing.T) {
	var buffer bytes.Buffer
	l := slog.New(slog.NewTextHandler(&buffer, &slog.HandlerOptions{ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
		if a.Key == slog.TimeKey {
			return slog.Attr{}
		}
		return a
	}}))

	probe := Probe{
		Format: Format{
			BitRate:  "1024000",
			Duration: "3600",
		},
		Streams: []Stream{
			{CodecType: "audio", CodecName: "aac"},
			{CodecType: "video", CodecName: "hevc", Height: 720},
		},
	}

	l.Info("probe", "probe", probe)
	assert.Equal(t, "level=INFO msg=probe probe.codec=hevc probe.bitrate=1000 probe.height=720 probe.duration=1h0m0s\n", buffer.String())
}
