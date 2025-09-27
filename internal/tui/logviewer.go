package tui

import (
	"codeberg.org/clambin/bubbles/frame"
	"codeberg.org/clambin/bubbles/stream"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var _ tea.Model = logViewer{}

type logViewer struct {
	tea.Model
	frameStyles frame.Styles
}

func (l logViewer) Init() tea.Cmd {
	return l.Model.Init()
}

func (l logViewer) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case stream.SetSizeMsg:
		msg.Width = max(0, msg.Width-l.frameStyles.Border.GetHorizontalBorderSize())
		msg.Height = max(0, msg.Height-l.frameStyles.Border.GetVerticalBorderSize())
		l.Model, cmd = l.Model.Update(msg)
	default:
		l.Model, cmd = l.Model.Update(msg)
	}
	return l, cmd
}

func (l logViewer) View() string {
	return frame.Draw("logs", lipgloss.Center, l.Model.View(), l.frameStyles)
}
