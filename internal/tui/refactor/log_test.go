package refactor

import (
	"testing"
	"time"

	"codeberg.org/clambin/bubbles/colors"
	"codeberg.org/clambin/bubbles/frame"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogViewer(t *testing.T) {
	l := NewLogViewer(DefaultLogViewerKeyMap(), frame.Styles{
		Title:  lipgloss.NewStyle().Foreground(lipgloss.Color("#ffffff")),
		Border: lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).Foreground(colors.Aqua),
	})
	l.SetSize(128, 4)

	go func() {
		cmd := l.Init()
		if cmd != nil {
			sendAndWait(l, cmd())
		}
	}()

	_, err := l.LogWriter().Write([]byte("Hello World\nNice to see you\n"))
	require.NoError(t, err)

	for {
		content := l.View()
		if content != "" {
			break
		}
		select {
		case <-t.Context().Done():
			return
		case <-time.After(time.Second):
		}
	}

	requireEqual(t, l.View())

	// esc closes the logViewer
	cmd := l.Update(tea.KeyMsg{Type: tea.KeyEsc})
	require.NotNil(t, cmd)
	assert.IsType(t, LogViewerClosedMsg{}, cmd())
}
