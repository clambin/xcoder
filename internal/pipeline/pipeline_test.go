package pipeline

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/clambin/videoConvertor/internal/configuration"
	"github.com/clambin/videoConvertor/internal/profile"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRun(t *testing.T) {
	tmpDir := t.TempDir()

	p, _ := profile.GetProfile("hevc-high")
	cfg := configuration.Configuration{
		Profile: p,
		Input:   tmpDir,
	}
	var queue Queue
	l := slog.New(slog.DiscardHandler)
	var errCh = make(chan error)
	ctx, cancel := context.WithCancel(t.Context())
	go func() { errCh <- Run(ctx, cfg, &queue, l) }()

	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "video.mkv"), []byte{}, 0644))

	assert.Eventually(t, func() bool { return len(queue.List()) > 0 }, time.Second, time.Millisecond)

	cancel()
	require.NoError(t, <-errCh)

	items := queue.List()
	require.Len(t, items, 1)
	status, _ := items[0].Status()
	assert.Equal(t, Failed, status)
}
