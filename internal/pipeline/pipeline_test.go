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
	p, _ := profile.GetProfile("hevc-high")
	cfg := configuration.Configuration{
		Profile: p,
		Input:   t.TempDir(),
	}
	var queue Queue
	l := slog.New(slog.DiscardHandler)
	var errCh = make(chan error)
	ctx, cancel := context.WithCancel(t.Context())
	go func() { errCh <- Run(ctx, cfg, &queue, l) }()

	require.NoError(t, os.WriteFile(filepath.Join(cfg.Input, "video.mkv"), []byte{}, 0644))

	assert.Eventually(t, func() bool {
		items := queue.List()
		if len(queue.List()) != 1 {
			return false
		}
		status, _ := items[0].Status()
		return status == Failed
	}, time.Second, time.Millisecond)

	cancel()
	require.NoError(t, <-errCh)
}
