package refactor

import (
	"io"

	"codeberg.org/clambin/bubbles/frame"
	"codeberg.org/clambin/bubbles/stream"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// LogViewer displays the log/slog output
type LogViewer struct {
	tea.Model
	frameStyles frame.Styles
	keyMap      LogViewerKeyMap
}

func NewLogViewer(keyMap LogViewerKeyMap, style frame.Styles) *LogViewer {
	return &LogViewer{
		Model: stream.NewStream(80, 25,
			stream.WithShowToggles(true),
			stream.WithKeyMap(stream.KeyMap{WordWrap: keyMap.WordWrap, AutoScroll: keyMap.AutoScroll}),
		),
		frameStyles: style,
		keyMap:      keyMap,
	}
}

func (l *LogViewer) Init() tea.Cmd {
	return l.Model.Init()
}

func (l *LogViewer) SetSize(width, height int) {
	// TODO: stream should just have a SetSize() method
	l.Model, _ = l.Model.Update(stream.SetSizeMsg{
		Width:  max(0, width-l.frameStyles.Border.GetHorizontalBorderSize()),
		Height: max(0, height-l.frameStyles.Border.GetVerticalBorderSize()),
	})
}

func (l *LogViewer) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, l.keyMap.CloseLogs):
			return func() tea.Msg { return LogViewerClosedMsg{} }
		default:
			l.Model, cmd = l.Model.Update(msg)
		}
	default:
		l.Model, cmd = l.Model.Update(msg)
	}
	return cmd
}

func (l *LogViewer) View() string {
	return frame.Draw("logs", lipgloss.Center, l.Model.View(), l.frameStyles)
}

func (l *LogViewer) LogWriter() io.Writer {
	return l.Model.(io.Writer)
}
