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
	"github.com/charmbracelet/x/exp/teatest"
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

func TestQueueViewer(t *testing.T) {
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

	qv := newQueueViewer(&fakeQueue{queue: worklist}, QueueViewerStyles{}, DefaultQueueViewerKeyMap())
	// queueViewer doesn't react to tea.WindowSizeMsg, so we need to set a size
	qv.SetSize(150, 10)

	// initialize a test model and wait for the screen to render
	tm := teatest.NewTestModel(t, queueViewWrapper{qv}, teatest.WithInitialTermSize(150, 10))
	tm.Send(refreshUIMsg{})
	waitFor(t, tm.Output(), []byte("converted"))
	golden.RequireEqual(t, qv.View())

	// test each of the media filters
	for _, r := range []rune{'x', 'r', 's', 'c'} {
		t.Run(string(r), func(t *testing.T) {
			// user changes filter
			sendAndWait(qv, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
			// check screen is updated
			golden.RequireEqual(t, qv.View())
		})
	}

}

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
	qv.SetSize(150, 10)

	// initialize the test model
	tm := teatest.NewTestModel(t, queueViewWrapper{qv}, teatest.WithInitialTermSize(150, 10))
	tm.Send(refreshUIMsg{})
	waitFor(t, tm.Output(), []byte("inspected"))

	// ask to convert the first item
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
	assert.Eventually(t, func() bool { return q.queue[0].WorkStatus().Status == pipeline.Converting }, time.Second, 10*time.Millisecond)

	// fullPath
	assert.False(t, qv.showFullPath)
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}})
	assert.Eventually(t, func() bool { return qv.showFullPath }, time.Second, 10*time.Millisecond)

	// activate queue
	assert.False(t, q.Active())
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	assert.Eventually(t, q.Active, time.Second, 10*time.Millisecond)

	// switch on text filter
	assert.False(t, qv.textFilterOn)
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	assert.Eventually(t, func() bool { return qv.textFilterOn }, time.Second, 10*time.Millisecond)

	// type some text
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	tm.Send(tea.KeyMsg{Type: tea.KeyEsc})
	assert.Eventually(t, func() bool { return !qv.textFilterOn }, time.Second, 10*time.Millisecond)
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

var _ tea.Model = queueViewWrapper{}

type queueViewWrapper struct {
	qvw *queueViewer
}

func (q queueViewWrapper) Init() tea.Cmd {
	return q.qvw.Init()
}

func (q queueViewWrapper) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return q, q.qvw.Update(msg)
}

func (q queueViewWrapper) View() string {
	return q.qvw.View()
}
