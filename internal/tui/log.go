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
	frameStyles frame.Styles
	keyMap      LogViewerKeyMap
}

func newLogViewer(keyMap LogViewerKeyMap, style LogViewerStyles) *logViewer {
	return &logViewer{
		Model: stream.NewStream(80, 25,
			stream.WithShowToggles(true),
			stream.WithKeyMap(stream.KeyMap{WordWrap: keyMap.WordWrap, AutoScroll: keyMap.AutoScroll}),
		),
		frameStyles: style.Frame,
		keyMap:      keyMap,
	}
}

func (l *logViewer) Init() tea.Cmd {
	return l.Model.Init()
}

func (l *logViewer) SetSize(width, height int) {
	// TODO: stream should just have a SetSize() method
	l.Model, _ = l.Model.Update(stream.SetSizeMsg{
		Width:  max(0, width-l.frameStyles.Border.GetHorizontalBorderSize()),
		Height: max(0, height-l.frameStyles.Border.GetVerticalBorderSize()),
	})
}

func (l *logViewer) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, l.keyMap.CloseLogs):
			return func() tea.Msg { return logViewerClosedMsg{} }
		default:
			l.Model, cmd = l.Model.Update(msg)
		}
	default:
		l.Model, cmd = l.Model.Update(msg)
	}
	return cmd
}

func (l *logViewer) View() string {
	return frame.Draw("logs", lipgloss.Center, l.Model.View(), l.frameStyles)
}

func (l *logViewer) LogWriter() io.Writer {
	return l.Model.(io.Writer)
}
