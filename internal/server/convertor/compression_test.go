package convertor

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
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

	l.Info("compression", "factor", CompressionFactor(0.12345))
	assert.Equal(t, "level=INFO msg=compression factor=0.12\n", output.String())
}

func TestCalculateCompression(t *testing.T) {
	tests := []struct {
		name    string
		source  int
		target  int
		wantErr assert.ErrorAssertionFunc
		want    CompressionFactor
	}{
		{
			name:    "smaller",
			source:  100,
			target:  10,
			wantErr: assert.NoError,
			want:    0.1,
		},
		{
			name:    "larger",
			source:  10,
			target:  100,
			wantErr: assert.NoError,
			want:    10,
		},
		{
			name:    "source missing",
			source:  0,
			target:  10,
			wantErr: assert.Error,
		},
		{
			name:    "target missing",
			source:  10,
			target:  0,
			wantErr: assert.Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpDir, err := os.MkdirTemp("", "")
			require.NoError(t, err)
			defer func() { assert.NoError(t, os.RemoveAll(tmpDir)) }()

			fullSource := filepath.Join(tmpDir, "source")
			if tt.source > 0 {
				content := strings.Repeat("A", tt.source)
				require.NoError(t, os.WriteFile(fullSource, []byte(content), 0644))
			}
			fullTarget := filepath.Join(tmpDir, "target")
			if tt.target > 0 {
				content := strings.Repeat("A", tt.target)
				require.NoError(t, os.WriteFile(fullTarget, []byte(content), 0644))
			}

			compression, err := CalculateCompression(fullSource, fullTarget)
			tt.wantErr(t, err)
			if err == nil {
				assert.Equal(t, tt.want, compression)
			}
		})
	}
}
