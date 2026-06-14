package ui

import (
	"io"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"codeberg.org/clambin/bubbles/frame"
	"codeberg.org/clambin/bubbles/helper"
	"codeberg.org/clambin/bubbles/stream"
)

// logViewer displays the log/slog output
type logViewer struct {
	style  LogViewerStyles
	keyMap LogViewerKeyMap
	stream.Model
}

func newLogViewer(r io.Reader, keyMap LogViewerKeyMap, style LogViewerStyles) logViewer {
	return logViewer{
		Model: stream.New(r,
			stream.WithShowToggles(true),
			stream.WithKeyMap(stream.KeyMap{WordWrap: keyMap.WordWrap, AutoScroll: keyMap.AutoScroll}),
		),
		style:  style,
		keyMap: keyMap,
	}
}

func (l logViewer) Update(msg tea.Msg) (logViewer, tea.Cmd) {
	var cmd tea.Cmd
	l.Model, cmd = l.Model.Update(msg)
	return l, cmd
}

func (l logViewer) View() string {
	content := l.style.Text.Render(l.Model.View())
	return frame.Render("logs", lipgloss.Center, l.style.Frame, content)
}

func (l logViewer) SetSize(width, height int) logViewer {
	borderWidth, borderHeight := l.style.Frame.BorderSize()
	l.Model = l.Model.SetSize(
		max(0, width-borderWidth),
		max(0, height-borderHeight),
	)
	return l
}

func (l logViewer) helpSections() []helper.Section {
	return []helper.Section{{Title: "LOGS", Keys: l.keyMap.FullHelp()[0]}}
}
