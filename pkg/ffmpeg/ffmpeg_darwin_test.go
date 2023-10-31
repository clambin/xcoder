package ffmpeg

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"log/slog"
	"testing"
)

func Test_runCommand(t *testing.T) {
	p := Processor{Logger: slog.Default()}
	output, err := p.runCommand(context.Background(), "bash", "-c", "echo foo")
	require.NoError(t, err)
	assert.Equal(t, "foo\n", output.String())

	_, err = p.runCommand(context.Background(), "invalid")
	assert.Error(t, err)
}
