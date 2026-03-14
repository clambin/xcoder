package tui

import (
	"slices"
	"sync/atomic"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/exp/teatest/v2"
	"github.com/clambin/xcoder/ffmpeg"
	"github.com/clambin/xcoder/internal/pipeline"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	qv = qv.SetSize(150, 10)

	// initialize the test model
	tm := teatest.NewTestModel(t, app[queueViewer]{model: qv}, teatest.WithInitialTermSize(150, 10))
	tm.Send(refreshUIMsg{})
	waitFor(t, tm.Output(), []byte("inspected"))

	// ask to convert the first item
	tm.Send(tea.KeyPressMsg{Code: tea.KeyEnter})
	assert.Eventually(t, func() bool { return q.queue[0].WorkStatus().Status == pipeline.Converting }, time.Second, 10*time.Millisecond)

	// fullPath
	tm.Send(tea.KeyPressMsg{Text: "f"})

	// activate queue
	tm.Send(tea.KeyPressMsg{Text: "p"})

	// switch on text filter
	assert.False(t, qv.textFilterOn)
	tm.Send(tea.KeyPressMsg{Text: "/"})

	// shut down to avoid race conditions reading status
	require.NoError(t, tm.Quit())

	// check results of actions
	assert.True(t, q.Active())
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
