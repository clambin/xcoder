package tui

import (
	"io"

	"codeberg.org/clambin/bubbles/frame"
	"codeberg.org/clambin/bubbles/stream"
	"codeberg.org/clambin/bubbles/table"
	"github.com/charmbracelet/bubbles/help"
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

// activePane determines which pane is active, i.e., which one gets keyboard input and which one to display.
type activePane int

const (
	queuePane activePane = iota
	logPane
)

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Queue viewer displays the queue of media files
//
// Since queueViewer does not have access to the selected media filters (converted, skipped, etc.),
// queueViewer is only responsible for displaying the table, keyboard input and local settings (like whether
// to show the full path or not).  More advanced functionality (like filtering) is done in the Controller.

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

type queueViewer struct {
	tea.Model
	keyMap       queueViewerKeyMap
	active       bool
	textFilterOn bool
	showFullPath bool
}

func newQueueViewer(keyMap queueViewerKeyMap, styles table.FilterTableStyles) queueViewer {
	return queueViewer{
		Model: table.NewFilterTable(
			"media files",
			columns,
			nil,
			styles,
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

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Log viewer displays the log/slog output

var _ help.KeyMap = logViewerKeyMap{}

type logViewerKeyMap struct {
	WordWrap   key.Binding
	AutoScroll key.Binding
	CloseLogs  key.Binding
}

func defaultLogViewerKeyMap() logViewerKeyMap {
	return logViewerKeyMap{
		WordWrap:   key.NewBinding(key.WithKeys("w"), key.WithHelp("w", "wrap words")),
		AutoScroll: key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "auto scroll")),
		CloseLogs:  key.NewBinding(key.WithKeys(tea.KeyEscape.String(), "l"), key.WithHelp(tea.KeyEscape.String()+"/l", "close logs")),
	}
}

func (l logViewerKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{l.CloseLogs, l.AutoScroll, l.CloseLogs}
}

func (l logViewerKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{l.ShortHelp()}
}

// logViewer displays the log/slog output
type logViewer struct {
	tea.Model
	frameStyles frame.Styles
	keyMap      logViewerKeyMap
	active      bool
}

func newLogViewer(keyMap logViewerKeyMap, style frame.Styles) logViewer {
	return logViewer{
		Model: stream.NewStream(80, 25,
			stream.WithShowToggles(true),
			stream.WithKeyMap(stream.KeyMap{WordWrap: keyMap.WordWrap, AutoScroll: keyMap.AutoScroll}),
		),
		frameStyles: style,
		keyMap:      keyMap,
	}
}

func (l logViewer) Update(msg tea.Msg) (logViewer, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case stream.SetSizeMsg:
		msg.Width = max(0, msg.Width-l.frameStyles.Border.GetHorizontalBorderSize())
		msg.Height = max(0, msg.Height-l.frameStyles.Border.GetVerticalBorderSize())
		l.Model, cmd = l.Model.Update(msg)
		return l, cmd
	case setPaneMsg:
		l.active = activePane(msg) == logPane
		return l, nil
	case tea.KeyMsg:
		if !l.active {
			break
		}
		switch {
		case key.Matches(msg, l.keyMap.CloseLogs):
			return l, setPaneCmd(queuePane)
		default:
			l.Model, cmd = l.Model.Update(msg)
		}
	default:
		l.Model, cmd = l.Model.Update(msg)
	}

	return l, cmd
}

func (l logViewer) View() string {
	return frame.Draw("logs", lipgloss.Center, l.Model.View(), l.frameStyles)
}

func (l logViewer) LogWriter() io.Writer {
	return l.Model.(io.Writer)
}

func (l logViewer) setSizeCmd(width, height int) tea.Cmd {
	return func() tea.Msg { return stream.SetSizeMsg{Width: width, Height: height} }
}
