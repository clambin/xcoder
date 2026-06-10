package transcoder

import (
	"fmt"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStatus_String(t *testing.T) {
	for status, val := range statusStrings {
		assert.Equal(t, val, status.String())
	}
	assert.Equal(t, "unknown", Status(-1).String())
}

func TestWorkItems(t *testing.T) {
	items := []*WorkItem{
		{Source: File{Path: "file1.mp4"}, status: StatusScanned},
		{Source: File{Path: "file2.mp4"}, status: StatusSkipped},
		{Source: File{Path: "file3.mp4"}, status: StatusRejected},
	}
	t.Run("add", func(t *testing.T) {
		var q WorkItems
		q.Add(items...)
		list := q.Items()
		assert.Equal(t, []string{"file1.mp4", "file2.mp4", "file3.mp4"}, reduce(list, func(item *WorkItem) string { return item.Source.Path }))
		assert.Equal(t, []Status{StatusScanned, StatusSkipped, StatusRejected}, reduce(list, func(item *WorkItem) Status { return item.status }))
	})
	t.Run("add duplicate", func(t *testing.T) {
		var q WorkItems
		q.Add(items...)
		q.Add(&WorkItem{Source: File{Path: "file1.mp4"}, status: StatusQueued})
		list := q.Items()
		assert.Equal(t, []string{"file1.mp4", "file2.mp4", "file3.mp4"}, reduce(list, func(item *WorkItem) string { return item.Source.Path }))
		assert.Equal(t, []Status{StatusQueued, StatusSkipped, StatusRejected}, reduce(list, func(item *WorkItem) Status { return item.status }))
	})
	t.Run("remove", func(t *testing.T) {
		var q WorkItems
		q.Add(items...)
		q.Remove(items[0])
		list := q.Items()
		assert.Equal(t, []string{"file2.mp4", "file3.mp4"}, reduce(list, func(item *WorkItem) string { return item.Source.Path }))
		assert.Equal(t, []Status{StatusSkipped, StatusRejected}, reduce(list, func(item *WorkItem) Status { return item.status }))
	})
	t.Run("get first", func(t *testing.T) {
		var q WorkItems
		q.Add(items...)
		// successful GetFirst
		item, ok := q.GetFirst(StatusSkipped)
		require.True(t, ok)
		assert.Equal(t, items[1], item)
		// unsuccessful GetFirst
		_, ok = q.GetFirst(StatusConverted)
		require.False(t, ok)

	})
	t.Run("items with status", func(t *testing.T) {
		var q WorkItems
		q.Add(items...)
		list := q.ItemsWithStatus(StatusSkipped)
		assert.Equal(t, []string{"file2.mp4"}, reduce(list, func(item *WorkItem) string { return item.Source.Path }))
	})
}

func BenchmarkWorkItems(b *testing.B) {
	const size = 1_000
	items := make([]*WorkItem, size)
	for i := range size {
		items[i] = &WorkItem{Source: File{Path: fmt.Sprintf("file%d.mp4", i)}, status: StatusQueued}
	}
	slices.Reverse(items)

	b.Run("add", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()
		for b.Loop() {
			var q WorkItems
			q.Add(items...)
			_ = q.Items()
		}
	})
	b.Run("add duplicate", func(b *testing.B) {
		var q WorkItems
		q.Add(items...)
		b.ResetTimer()
		b.ReportAllocs()
		for b.Loop() {
			q.Add(items[0])
			_ = q.Items()
		}
	})
}

func reduce[T []E, E any, V any](slice T, f func(E) V) []V {
	result := make([]V, len(slice))
	for i, e := range slice {
		result[i] = f(e)
	}
	return result
}
