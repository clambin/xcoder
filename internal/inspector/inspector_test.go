package inspector

import (
	"context"
	"github.com/clambin/vidconv/internal/feeder"
	"github.com/clambin/vidconv/internal/inspector/mocks"
	"github.com/clambin/vidconv/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"log/slog"
	"testing"
	"time"
)

func TestInspector_Run(t *testing.T) {
	vp := mocks.NewVideoProcessor(t)
	stats := testutil.MakeProbe("h264", 5*1024, 720, time.Hour)
	vp.EXPECT().Probe(mock.Anything, mock.AnythingOfType("string")).Return(stats, nil)

	ch := make(chan feeder.Entry)
	i := New(ch, "hevc", slog.Default())
	i.VideoProcessor = vp

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error)
	go func() {
		errCh <- i.Run(ctx)
	}()

	ch <- feeder.Entry{
		Path: "/foo/bar.mkv",
		DirEntry: fakeDirEntry{
			name:    "bar.mkv",
			isDir:   false,
			modTime: time.Date(2023, time.November, 7, 0, 0, 0, 0, time.UTC),
		},
	}

	assert.Equal(t, Video{
		Path:    "/foo/bar.mkv",
		ModTime: time.Date(2023, time.November, 7, 0, 0, 0, 0, time.UTC),
		Info:    VideoInfo{Name: "bar", Extension: "mkv", IsSeries: false, Episode: ""},
		Stats:   stats,
	}, <-i.Output)
	cancel()
	assert.NoError(t, <-errCh)
}
