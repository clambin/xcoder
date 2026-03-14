package tui

import (
	"testing"

	"charm.land/lipgloss/v2"
	"codeberg.org/clambin/bubbles/colors"
	"codeberg.org/clambin/bubbles/frame"
	"github.com/charmbracelet/x/exp/teatest/v2"
)

func TestLogViewer(t *testing.T) {
	l := newLogViewer(DefaultLogViewerKeyMap(), LogViewerStyles{Frame: frame.Style{
		Title:  lipgloss.NewStyle().Foreground(lipgloss.Color("#ffffff")),
		Border: lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).Foreground(colors.Aqua),
	}}).SetSize(128, 4)

	tm := teatest.NewTestModel(t, app[logViewer]{l})

	go func() {
		_, _ = l.Write([]byte("Hello World\nNice to see you\n"))
	}()
	waitFor(t, tm.Output(), []byte("Nice to see you"))
}
