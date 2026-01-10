package tui

import (
	"codeberg.org/clambin/bubbles/table"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	columns = []table.Column{
		{Name: "SOURCE"},
		{Name: "SOURCE STATS", Width: 25},
		{Name: "TARGET STATS", Width: 25},
		{Name: "STATUS", Width: 10, RowStyle: table.CellStyle{Style: lipgloss.NewStyle().Transform(table.StringStyler(statusColors))}},
		{Name: "COMPLETED", Width: 10, RowStyle: table.CellStyle{Style: lipgloss.NewStyle().Align(lipgloss.Right)}},
		{Name: "REMAINING", Width: 10, RowStyle: table.CellStyle{Style: lipgloss.NewStyle().Align(lipgloss.Right)}},
		{Name: "ERROR", Width: 40},
	}
)

// queueViewer displays the queue of media files
//
// Since queueViewer does not have access to the selected media filters (converted, skipped, etc.),
// queueViewer is only responsible for displaying the table, keyboard input, and local settings (like whether
// to show the full path or not).  More advanced functionality (like filtering) is done in the Controller.
type queueViewer struct {
	tea.Model
	keyMap       queueViewerKeyMap
	textFilterOn bool
	showFullPath bool
}

func newQueueViewer(keyMap queueViewerKeyMap, tableStyles table.FilterTableStyles) tea.Model {
	return queueViewer{
		Model: table.NewFilterTable(
			"media files",
			columns,
			nil,
			tableStyles,
			keyMap.FilterTableKeyMap,
		),
		keyMap: keyMap,
	}
}

func (q queueViewer) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case paneSizeMsg:
		q.Model, cmd = q.Model.Update(table.SetSizeMsg{Width: msg.Width, Height: msg.Height})
	case table.FilterStateChangeMsg:
		q.textFilterOn = msg.State
		cmd = textFilterStateChangCmd(q.textFilterOn)
	case tea.KeyMsg:
		// if the text mediaFilter is active, it receives all inputs.
		if q.textFilterOn {
			q.Model, cmd = q.Model.Update(msg)
			break
		}
		switch {
		case key.Matches(msg, q.keyMap.FullPath):
			q.showFullPath = !q.showFullPath
			// note: don't issue table reload here: msgs aren't guaranteed to be processed in order
			cmd = showFullPathCmd(q.showFullPath)
		default:
			q.Model, cmd = q.Model.Update(msg)
		}
	default:
		q.Model, cmd = q.Model.Update(msg)
	}

	return q, cmd
}

func (q queueViewer) View() string {
	return q.Model.View()
}

func (q queueViewer) TextFilterOn() bool {
	return q.textFilterOn
}
