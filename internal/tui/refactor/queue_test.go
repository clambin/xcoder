package refactor

import (
	"errors"
	"fmt"
	"iter"
	"sync/atomic"
	"testing"
	"time"

	"codeberg.org/clambin/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/clambin/xcoder/ffmpeg"
	"github.com/clambin/xcoder/internal/pipeline"
	"github.com/muesli/termenv"
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

func init() {
	lipgloss.SetColorProfile(termenv.ANSI)
}

func TestManager_Actions(t *testing.T) {
	worklist := []*pipeline.WorkItem{
		{
			Source: pipeline.MediaFile{Path: "test.mp4", VideoStats: h264VideoStats},
			Target: pipeline.MediaFile{Path: "test.hevc.mkv", VideoStats: hevcVideoStats},
		},
	}
	worklist[0].SetWorkStatus(pipeline.WorkStatus{Status: pipeline.Inspected})

	q := fakeQueue{queue: worklist}
	mgr := NewQueueViewer(&q, table.FilterTableStyles{}, DefaultKeyMap())
	sendAndWait(mgr, RefreshUIMsg{})

	// initialize model: wait until the table is loaded and the selectedRow is set.
	sendAndWait(mgr, tea.KeyMsg{Type: tea.KeyEnter})
	assert.Equal(t, pipeline.Converting, q.queue[0].WorkStatus().Status)

	// fullPath
	assert.False(t, mgr.showFullPath)
	sendAndWait(mgr, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}})
	assert.True(t, mgr.showFullPath)

	// activate queue
	assert.False(t, q.Active())
	sendAndWait(mgr, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	assert.True(t, q.Active())

	// switch on text filter
	assert.False(t, mgr.textFilterOn)
	sendAndWait(mgr, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	assert.True(t, mgr.textFilterOn)
	// type some text
	sendAndWait(mgr, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	sendAndWait(mgr, tea.KeyMsg{Type: tea.KeyEsc})
	assert.False(t, mgr.textFilterOn)
}

func TestManager_Filter(t *testing.T) {
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

	mgr := NewQueueViewer(&q, table.FilterTableStyles{}, DefaultKeyMap())
	// QueueViewer's table needs a size, or it doesn't render
	mgr.SetSize(256, 10)
	sendAndWait(mgr, RefreshUIMsg{})

	for _, r := range []rune{'x', 'r', 's', 'c'} {
		t.Run(string(r), func(t *testing.T) {
			sendAndWait(mgr, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
			requireEqual(t, mgr.View())
		})
	}
}

var _ Queue = (*fakeQueue)(nil)

type fakeQueue struct {
	queue  []*pipeline.WorkItem
	active atomic.Bool
}

func (f *fakeQueue) Stats() map[pipeline.Status]int {
	return map[pipeline.Status]int{
		pipeline.Inspected: 1,
	}
}

func (f *fakeQueue) All() iter.Seq[*pipeline.WorkItem] {
	return func(yield func(*pipeline.WorkItem) bool) {
		for _, item := range f.queue {
			if !yield(item) {
				return
			}
		}
	}
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
