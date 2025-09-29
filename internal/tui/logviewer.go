package tui

import (
	"io"

	"codeberg.org/clambin/bubbles/frame"
	"codeberg.org/clambin/bubbles/stream"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

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

var _ help.KeyMap = logViewerKeyMap{}

// logViewerKeyMap contains the key bindings for the logViewer.
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
