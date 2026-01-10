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
	tm := teatest.NewTestModel(t, q)
	tm.Send(table.SetSizeMsg{Width: 200, Height: 6}) // set the size of the log pane
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
