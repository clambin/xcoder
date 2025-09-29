package tui

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"codeberg.org/clambin/bubbles/stream"
	"codeberg.org/clambin/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/exp/teatest"
	"github.com/clambin/xcoder/internal/pipeline"
)

func TestQueueViewer(t *testing.T) {
	q := newQueueViewer(defaultQueueViewerKeyMap(), DefaultStyles().TableStyle)
	tm := teatest.NewTestModel(t, wrapper[queueViewer]{q})
	tm.Send(table.SetSizeMsg{Width: 200, Height: 6}) // set the size of the log pane
	tm.Send(setPaneMsg(queuePane))                   // activate the log pane
	rows := []table.Row{
		{"source", "source stats", "target stats", "waiting", "", "", "", table.UserData{Data: &pipeline.WorkItem{}}},
	}
	tm.Send(table.SetRowsMsg{Rows: rows})

	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return bytes.Contains(bts, []byte("waiting"))
	}, teatest.WithDuration(time.Second), teatest.WithCheckInterval(10*time.Millisecond),
	)

	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})

	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return strings.Contains(ansi.Strip(string(bts)), "filter")
	}, teatest.WithDuration(time.Second), teatest.WithCheckInterval(10*time.Millisecond),
	)
}

func TestLogViewer(t *testing.T) {
	l := newLogViewer(
		defaultLogViewerKeyMap(),
		DefaultStyles().FrameStyle,
	)
	_, _ = l.LogWriter().Write([]byte("Hello World\nNice to see you\n"))
	tm := teatest.NewTestModel(t, wrapper[logViewer]{l})
	tm.Send(stream.SetSizeMsg{Width: 40, Height: 6})            // set the size of the log pane
	tm.Send(setPaneMsg(logPane))                                // activate the log pane
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'w'}}) // switch on word wrap

	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return bytes.Contains(bts, []byte("Nice to see you"))
	}, teatest.WithDuration(time.Second), teatest.WithCheckInterval(10*time.Millisecond),
	)
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// a bit of generics magic to turn Update(tea.Msg) (myModel, tea.Cmd) into Update(tea.Msg) (tea.Model, tea.Cmd)

type adaptableModel[T any] interface {
	Init() tea.Cmd
	Update(tea.Msg) (T, tea.Cmd)
	View() string
}

var _ tea.Model = wrapper[logViewer]{}

type wrapper[T adaptableModel[T]] struct {
	t T
}

func (w wrapper[T]) Init() tea.Cmd {
	return w.t.Init()
}

func (w wrapper[T]) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	w.t, cmd = w.t.Update(msg)
	return w, cmd
}

func (w wrapper[T]) View() string {
	return w.t.View()
}
