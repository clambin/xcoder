package tui

import (
	"io"

	"codeberg.org/clambin/bubbles/frame"
	"codeberg.org/clambin/bubbles/stream"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var _ tea.Model = logViewer{}

// logViewer displays the log/slog output
type logViewer struct {
	tea.Model
	frameStyles frame.Styles
	keyMap      logViewerKeyMap
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

func (l logViewer) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case paneSizeMsg:
		l.Model, cmd = l.Model.Update(stream.SetSizeMsg{
			Width:  max(0, msg.Width-l.frameStyles.Border.GetHorizontalBorderSize()),
			Height: max(0, msg.Height-l.frameStyles.Border.GetVerticalBorderSize()),
		})
		return l, cmd
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, l.keyMap.CloseLogs):
			return l, showLogsCmd(false)
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
