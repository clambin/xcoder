package processor

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"log/slog"
	"testing"
)

func TestCompressionFactor_LogValue(t *testing.T) {
	var output bytes.Buffer
	l := slog.New(slog.NewTextHandler(&output, &slog.HandlerOptions{ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
		if a.Key == slog.TimeKey {
			return slog.Attr{}
		}
		return a
	}}))

	l.Info("compression", "factor", compressionFactor(0.12345))
	assert.Equal(t, "level=INFO msg=compression factor=0.12\n", output.String())
}
