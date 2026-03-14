package tui

import (
	"io"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"codeberg.org/clambin/bubbles/frame"
	"codeberg.org/clambin/bubbles/stream"
)

// logViewer displays the log/slog output
type logViewer struct {
	frameStyle frame.Style
	keyMap     LogViewerKeyMap
	stream.Stream
}

func newLogViewer(keyMap LogViewerKeyMap, style LogViewerStyles) logViewer {
	return logViewer{
		Stream: stream.New(
			stream.WithShowToggles(true),
			stream.WithKeyMap(stream.KeyMap{WordWrap: keyMap.WordWrap, AutoScroll: keyMap.AutoScroll}),
		),
		frameStyle: style.Frame,
		keyMap:     keyMap,
	}
}

func (l logViewer) Update(msg tea.Msg) (logViewer, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, l.keyMap.CloseLogs):
			return l, func() tea.Msg { return logViewerClosedMsg{} }
		}
	}
	var cmd tea.Cmd
	l.Stream, cmd = l.Stream.Update(msg)
	return l, cmd
}

func (l logViewer) View() string {
	return frame.Render("logs", lipgloss.Center, l.frameStyle, l.Model.View())
}

func (l logViewer) LogWriter() io.Writer {
	return l.Stream
}

func (l logViewer) SetSize(width, height int) logViewer {
	l.Stream = l.Size(
		max(0, width-l.frameStyle.Border.GetHorizontalBorderSize()),
		max(0, height-l.frameStyle.Border.GetVerticalBorderSize()),
	)
	return l
}
