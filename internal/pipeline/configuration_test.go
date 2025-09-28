package pipeline

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetConfigurationFromViper(t *testing.T) {
	v := viper.New()
	v.Set("profile", "invalid")

	_, err := GetConfigurationFromViper(v)
	require.Error(t, err)

	v.Set("profile", "hevc-high")
	cfg, err := GetConfigurationFromViper(v)
	require.NoError(t, err)
	assert.Equal(t, "hevc", cfg.Profile.TargetCodec)
}

func TestLog(t *testing.T) {
	var buf bytes.Buffer
	opts := slog.HandlerOptions{ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
		if a.Key == "time" {
			return slog.Attr{}
		}
		return a
	}}
	l := Log{Level: "info", Format: "json"}.Logger(&buf, &opts)
	l.Info("hello world")
	assert.Equal(t, `{"level":"INFO","msg":"hello world"}
`, buf.String())

	buf.Reset()
	l = Log{Level: "info", Format: "text"}.Logger(&buf, &opts)
	l.Info("hello world")
	assert.Equal(t, `level=INFO msg="hello world"
`, buf.String())

}
