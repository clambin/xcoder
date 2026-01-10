package tui

import (
	"bytes"
	"io"
	"iter"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/exp/teatest"
	"github.com/clambin/xcoder/ffmpeg"
	"github.com/clambin/xcoder/internal/pipeline"
	"github.com/muesli/termenv"
)

func init() {
	lipgloss.SetColorProfile(termenv.ANSI)
}

func TestController(t *testing.T) {
	worklist := []*pipeline.WorkItem{
		{
			Source: pipeline.MediaFile{
				Path: "test.mp4",
				VideoStats: ffmpeg.VideoStats{
					VideoCodec:    "h264",
					Duration:      time.Hour,
					BitRate:       10_000_000,
					BitsPerSample: 10,
					Height:        1080,
					Width:         2000,
				},
			},
			Target: pipeline.MediaFile{
				Path: "test.hevc.mkv",
				VideoStats: ffmpeg.VideoStats{
					VideoCodec:    "h264",
					Duration:      time.Hour,
					BitRate:       5_000_000,
					BitsPerSample: 10,
					Height:        1080,
					Width:         2000,
				},
			},
		},
	}
	worklist[0].SetWorkStatus(pipeline.WorkStatus{Status: pipeline.Inspected})

	q := fakeQueue{queue: worklist}
	c := New(&q, pipeline.Configuration{})
	tm := teatest.NewTestModel(t, c, teatest.WithInitialTermSize(128, 25))
	waitFor(t, tm.Output(), []byte("inspected"))

	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
	waitFor(t, tm.Output(), []byte("converting"))

	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlC})
	tm.WaitFinished(t)
}

func waitFor(t testing.TB, r io.Reader, want []byte) {
	t.Helper()
	waitForFunc(t, r, func(b []byte) bool { return bytes.Contains(b, want) })
}

func waitForFunc(t testing.TB, r io.Reader, f func([]byte) bool) {
	t.Helper()
	teatest.WaitFor(t, r, f, teatest.WithDuration(time.Second), teatest.WithCheckInterval(10*time.Millisecond))
}

var _ Queue = (*fakeQueue)(nil)

type fakeQueue struct {
	queue []*pipeline.WorkItem
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

func (f *fakeQueue) SetActive(_ bool) {
	panic("implement me")
}

func (f *fakeQueue) Active() bool {
	return false
}
