package tui

import (
	"io"

	"codeberg.org/clambin/bubbles/frame"
	"codeberg.org/clambin/bubbles/stream"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// logViewer displays the log/slog output
type logViewer struct {
	tea.Model
	frameStyle frame.Style
	keyMap     LogViewerKeyMap
}

func newLogViewer(keyMap LogViewerKeyMap, style LogViewerStyles) logViewer {
	return logViewer{
		Model: stream.New(
			stream.WithShowToggles(true),
			stream.WithKeyMap(stream.KeyMap{WordWrap: keyMap.WordWrap, AutoScroll: keyMap.AutoScroll}),
		),
		frameStyle: style.Frame,
		keyMap:     keyMap,
	}
}

func (l logViewer) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, l.keyMap.CloseLogs):
			return l, func() tea.Msg { return logViewerClosedMsg{} }
		}
	}
	var cmd tea.Cmd
	l.Model, cmd = l.Model.Update(msg)
	return l, cmd
}

func (l logViewer) View() string {
	return frame.Draw("logs", lipgloss.Center, l.Model.View(), l.frameStyle)
}

func (l logViewer) LogWriter() io.Writer {
	return l.Model.(stream.Stream)
}

func (l logViewer) SetSize(width, height int) logViewer {
	l.Model = l.Model.(stream.Stream).Size(
		max(0, width-l.frameStyle.Border.GetHorizontalBorderSize()),
		max(0, height-l.frameStyle.Border.GetVerticalBorderSize()),
	)
	return l
}
