package tui

import (
	"strings"
	"testing"

	"codeberg.org/clambin/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/exp/teatest"
	"github.com/clambin/xcoder/internal/pipeline"
)

func TestQueueViewer(t *testing.T) {
	q := newQueueViewer(defaultQueueViewerKeyMap(), DefaultStyles().TableStyle)
	tm := teatest.NewTestModel(t, queueViewerModel{q})
	tm.Send(table.SetSizeMsg{Width: 200, Height: 6}) // set the size of the log pane
	tm.Send(setPaneMsg(queuePane))                   // activate the log pane
	rows := []table.Row{
		{"source", "source stats", "target stats", "waiting", "", "", "", table.UserData{Data: &pipeline.WorkItem{}}},
	}
	tm.Send(table.SetRowsMsg{Rows: rows})
	waitFor(t, tm.Output(), []byte("waiting"))

	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	waitForFunc(t, tm.Output(), func(b []byte) bool {
		return strings.Contains(ansi.Strip(string(b)), "filter")
	})
}

var _ tea.Model = queueViewerModel{}

type queueViewerModel struct {
	queueViewer
}

func (m queueViewerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.queueViewer, cmd = m.queueViewer.Update(msg)
	return m, cmd
}
