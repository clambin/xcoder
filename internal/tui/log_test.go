package tui

import (
	"testing"

	"codeberg.org/clambin/bubbles/colors"
	"codeberg.org/clambin/bubbles/frame"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/exp/teatest"
)

func TestLogViewer2(t *testing.T) {
	l := newLogViewer(DefaultLogViewerKeyMap(), LogViewerStyles{Frame: frame.Styles{
		Title:  lipgloss.NewStyle().Foreground(lipgloss.Color("#ffffff")),
		Border: lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).Foreground(colors.Aqua),
	}})
	l.SetSize(128, 4)

	tm := teatest.NewTestModel(t, wrapper{l})

	go func() {
		_, _ = l.LogWriter().Write([]byte("Hello World\nNice to see you\n"))
	}()
	waitFor(t, tm.Output(), []byte("Nice to see you"))
}

var _ tea.Model = wrapper{}

type wrapper struct {
	c *logViewer
}

func (w wrapper) Init() tea.Cmd {
	return w.c.Init()
}

func (w wrapper) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return w, w.c.Update(msg)
}

func (w wrapper) View() string {
	return w.c.View()
}
