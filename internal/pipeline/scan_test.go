package pipeline

import (
	"log/slog"
	"slices"
	"testing"

	"github.com/psanford/memfs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScanFS(t *testing.T) {
	fs := memfs.New()
	require.NoError(t, fs.MkdirAll("foo", 0755))
	require.NoError(t, fs.WriteFile("foo/video.MKV", []byte(""), 0644))
	require.NoError(t, fs.WriteFile("foo/info.txt", []byte(""), 0644))
	require.NoError(t, fs.MkdirAll("bar", 0000))

	var queue Queue
	ch := make(chan *WorkItem)
	errCh := make(chan error)
	go func() { errCh <- ScanFS(t.Context(), fs, "/", &queue, ch, slog.Default()) }()
	item := <-ch
	assert.Equal(t, "/foo/video.MKV", item.Source.Path)
	require.NoError(t, <-errCh)

	items := slices.Collect(queue.All())
	require.Len(t, items, 1)
	assert.Equal(t, "/foo/video.MKV", items[0].Source.Path)
}
