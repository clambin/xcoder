package feeder_test

import (
	"context"
	"github.com/clambin/videoConvertor/internal/server/scanner/feeder"
	"github.com/clambin/videoConvertor/internal/testutil"
	"github.com/stretchr/testify/assert"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"
)

var validVideoFiles = []string{"foo.2021.avi", "foo.2021.mkv", "foo.2021.mp4"}

func TestFeeder_Run(t *testing.T) {
	h := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	tmpDir := testutil.MakeTempFS(t, validVideoFiles)
	defer func() { _ = os.RemoveAll(tmpDir) }()
	f := feeder.New(tmpDir, 200*time.Millisecond, slog.New(h))

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error)
	go func() { errCh <- f.Run(ctx) }()

	for i := 0; i < len(validVideoFiles); i++ {
		entry := <-f.Feed
		assert.Contains(t, validVideoFiles, filepath.Base(entry.Path))
	}

	cancel()
	assert.NoError(t, <-errCh)
}
