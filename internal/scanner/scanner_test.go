package scanner

import (
	"context"
	"github.com/clambin/videoConvertor/internal/worklist"
	"github.com/psanford/memfs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"log/slog"
	"testing"
)

func TestScanFS(t *testing.T) {

	ctx := context.Background()
	fs := memfs.New()

	require.NoError(t, fs.MkdirAll("foo", 0755))
	require.NoError(t, fs.WriteFile("foo/video.mkv", []byte(""), 0644))
	require.NoError(t, fs.WriteFile("foo/info.txt", []byte(""), 0644))
	require.NoError(t, fs.MkdirAll("bar", 0000))

	var list worklist.WorkList
	ch := make(chan *worklist.WorkItem)
	errCh := make(chan error)
	go func() { errCh <- ScanFS(ctx, fs, "/", &list, ch, slog.Default()) }()
	item := <-ch
	assert.Equal(t, "/foo/video.mkv", item.Source)
	assert.NoError(t, <-errCh)

	require.Len(t, list.List(), 1)
	assert.Equal(t, "/foo/video.mkv", list.List()[0].Source)
}
