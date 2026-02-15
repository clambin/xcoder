package tui

import (
	"testing"

	"codeberg.org/clambin/bubbles/colors"
	"codeberg.org/clambin/bubbles/frame"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/exp/teatest"
)

func TestLogViewer2(t *testing.T) {
	l := newLogViewer(DefaultLogViewerKeyMap(), LogViewerStyles{Frame: frame.Style{
		Title:  lipgloss.NewStyle().Foreground(lipgloss.Color("#ffffff")),
		Border: lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).Foreground(colors.Aqua),
	}})
	l.SetSize(128, 4)

	tm := teatest.NewTestModel(t, l)

	go func() {
		_, _ = l.LogWriter().Write([]byte("Hello World\nNice to see you\n"))
	}()
	waitFor(t, tm.Output(), []byte("Nice to see you"))
}
