package feeder_test

import (
	"context"
	"github.com/clambin/vidconv/internal/feeder"
	"github.com/clambin/vidconv/internal/feeder/mocks"
	"github.com/clambin/vidconv/internal/ffmpeg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFeeder_Run(t *testing.T) {
	h := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	tmpDir := makeTempFS(t)
	defer func() { _ = os.RemoveAll(tmpDir) }()
	s := feeder.New(tmpDir, 200*time.Millisecond, slog.New(h))

	vp := mocks.NewVideoProcessor(t)
	vp.EXPECT().Probe(mock.Anything, mock.AnythingOfType("string")).Return(ffmpeg.Probe{
		Streams: []ffmpeg.Stream{{CodecType: "video", CodecName: "hevc"}},
		Format: ffmpeg.Format{
			Duration: "1200.00",
		},
	}, nil)
	s.VideoProcessor = vp

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error)
	go func() { errCh <- s.Run(ctx) }()

	for i := 0; i < len(validVideoFiles); i++ {
		entry := <-s.Feed
		assert.Contains(t, validVideoFiles, filepath.Base(entry.Path))
		assert.Equal(t, "hevc", entry.Codec)
		assert.Equal(t, 1200*time.Second, entry.Duration)
	}

	cancel()
	assert.NoError(t, <-errCh)
}

var validVideoFiles = []string{"foo.2021.avi", "foo.2021.mkv", "foo.2021.mp4"}

func makeTempFS(t *testing.T) string {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "")
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "videofiles"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "videofiles/foo.2021.mkv"), nil, 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "videofiles/foo.2021.avi"), nil, 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "videofiles/foo.2021.mp4"), nil, 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "videofiles/foo.2021.srt"), nil, 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "videofiles/foo.2021.hid"), nil, 0000))
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "hidden"), 0000))

	return tmpDir
}
