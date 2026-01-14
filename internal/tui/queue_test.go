package tui

import (
	"errors"
	"fmt"
	"slices"
	"sync/atomic"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/golden"
	"github.com/clambin/xcoder/ffmpeg"
	"github.com/clambin/xcoder/internal/pipeline"
	"github.com/stretchr/testify/assert"
)

var (
	h264VideoStats = ffmpeg.VideoStats{
		VideoCodec:    "h264",
		Duration:      time.Hour,
		BitRate:       10_000_000,
		BitsPerSample: 10,
		Height:        1080,
		Width:         2000,
	}
	hevcVideoStats = ffmpeg.VideoStats{
		VideoCodec:    "hevc",
		Duration:      time.Hour,
		BitRate:       5_000_000,
		BitsPerSample: 10,
		Height:        1080,
		Width:         2000,
	}
)

func TestQueueViewer_Actions(t *testing.T) {
	worklist := []*pipeline.WorkItem{
		{
			Source: pipeline.MediaFile{Path: "test.mp4", VideoStats: h264VideoStats},
			Target: pipeline.MediaFile{Path: "test.hevc.mkv", VideoStats: hevcVideoStats},
		},
	}
	worklist[0].SetWorkStatus(pipeline.WorkStatus{Status: pipeline.Inspected})

	q := fakeQueue{queue: worklist}
	qv := newQueueViewer(&q, QueueViewerStyles{}, DefaultQueueViewerKeyMap())
	sendAndWait(qv, refreshUIMsg{})

	// initialize model: wait until the table is loaded and the selectedRow is set.
	sendAndWait(qv, tea.KeyMsg{Type: tea.KeyEnter})
	assert.Equal(t, pipeline.Converting, q.queue[0].WorkStatus().Status)

	// fullPath
	assert.False(t, qv.showFullPath)
	sendAndWait(qv, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}})
	assert.True(t, qv.showFullPath)

	// activate queue
	assert.False(t, q.Active())
	sendAndWait(qv, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	assert.True(t, q.Active())

	// switch on text filter
	assert.False(t, qv.textFilterOn)
	sendAndWait(qv, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	assert.True(t, qv.textFilterOn)
	// type some text
	sendAndWait(qv, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	sendAndWait(qv, tea.KeyMsg{Type: tea.KeyEsc})
	assert.False(t, qv.textFilterOn)
}

func TestQueueViewer_Filter(t *testing.T) {
	stats := []pipeline.Status{
		pipeline.Waiting,
		pipeline.Inspected,
		pipeline.Skipped,
		pipeline.Rejected,
		pipeline.Converting,
		pipeline.Failed,
		pipeline.Converted,
	}
	worklist := make([]*pipeline.WorkItem, len(stats))
	for i := range worklist {
		worklist[i] = &pipeline.WorkItem{
			Source: pipeline.MediaFile{Path: fmt.Sprintf("test-%d.mp4", i), VideoStats: h264VideoStats},
			Target: pipeline.MediaFile{Path: fmt.Sprintf("test-%d.hevc.mkv", i), VideoStats: hevcVideoStats},
		}
		var err error
		if stats[i] == pipeline.Failed {
			err = errors.New("test error")
		}
		worklist[i].SetWorkStatus(pipeline.WorkStatus{Status: stats[i], Err: err})
	}

	q := fakeQueue{queue: worklist}

	qv := newQueueViewer(&q, QueueViewerStyles{}, DefaultQueueViewerKeyMap())
	// queueViewer's table needs a size, or it doesn't render
	qv.SetSize(256, 10)

	for _, r := range []rune{'x', 'r', 's', 'c'} {
		t.Run(string(r), func(t *testing.T) {
			// user changes filter
			sendAndWait(qv, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
			// controller issues refreshUIMsg
			sendAndWait(qv, refreshUIMsg{})
			// check screen is updated
			golden.RequireEqual(t, qv.View())
		})
	}
}

var _ Queue = (*fakeQueue)(nil)

type fakeQueue struct {
	queue  []*pipeline.WorkItem
	active atomic.Bool
}

func (f *fakeQueue) Stats() map[pipeline.Status]int {
	stats := make(map[pipeline.Status]int)
	for _, item := range f.queue {
		stats[item.WorkStatus().Status]++
	}
	return stats
}

func (f *fakeQueue) All() []*pipeline.WorkItem {
	return slices.Clone(f.queue)
}

func (f *fakeQueue) Queue(*pipeline.WorkItem) {
	f.queue[0].SetWorkStatus(pipeline.WorkStatus{Status: pipeline.Converting})
}

func (f *fakeQueue) SetActive(active bool) {
	f.active.Store(active)
}

func (f *fakeQueue) Active() bool {
	return f.active.Load()
}
