package requests_test

import (
	"bytes"
	"github.com/clambin/videoConvertor/internal/server/requests"
	"github.com/clambin/videoConvertor/internal/testutil"
	"github.com/clambin/videoConvertor/pkg/ffmpeg"
	"github.com/stretchr/testify/assert"
	"log/slog"
	"testing"
	"time"
)

func TestConversionRequest_LogValue(t *testing.T) {
	var output bytes.Buffer
	l := slog.New(slog.NewJSONHandler(&output, &slog.HandlerOptions{ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
		if a.Key == slog.TimeKey {
			return slog.Attr{}
		}
		return a
	}}))

	r := requests.Request{
		Request: ffmpeg.Request{
			Source:        "/tmp/foo.mp4",
			Target:        "/tmp/foo.hevc.mp4",
			VideoCodec:    "hevc",
			BitRate:       1024000,
			BitsPerSample: 8,
		},
		SourceStats: testutil.MakeProbe("h264", 8, 720, time.Hour),
	}
	l.Info("log", slog.Any("request", r))
	assert.Equal(t, `{"level":"INFO","msg":"log","request":{"source":{"filename":"/tmp/foo.mp4","codec":"h264","bitrate":8192,"bitsPerSample":8},"target":{"filename":"/tmp/foo.hevc.mp4","codec":"hevc","bitrate":1024000,"bitsPerSample":8}}}
`, output.String())
}
