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
	active       bool
	textFilterOn bool
	showFullPath bool
}

func newQueueViewer(keyMap queueViewerKeyMap, tableStyles table.FilterTableStyles) queueViewer {
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

func (q queueViewer) Update(msg tea.Msg) (queueViewer, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case setPaneMsg:
		q.active = activePane(msg) == queuePane
	case table.FilterStateChangeMsg:
		q.textFilterOn = msg.State
	case tea.KeyMsg:
		if !q.active {
			break
		}
		// if the text filter is on, it receives all inputs.
		if q.textFilterOn {
			q.Model, cmd = q.Model.Update(msg)
			break
		}
		switch {
		case key.Matches(msg, q.keyMap.FullPath):
			q.showFullPath = !q.showFullPath
			cmd = func() tea.Msg { return refreshTableMsg{} }
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
	return q.active && q.textFilterOn
}

func (q queueViewer) setSizeCmd(width, height int) tea.Cmd {
	return func() tea.Msg { return table.SetSizeMsg{Width: width, Height: height} }
}

// queueViewerKeyMap contains the key bindings for the queueViewer.
type queueViewerKeyMap struct {
	FilterTableKeyMap table.FilterTableKeyMap
	FullPath          key.Binding
}

func defaultQueueViewerKeyMap() queueViewerKeyMap {
	return queueViewerKeyMap{
		FilterTableKeyMap: table.DefaultFilterTableKeyMap(),
		FullPath:          key.NewBinding(key.WithKeys("f"), key.WithHelp("f", "show full path")),
	}
}
