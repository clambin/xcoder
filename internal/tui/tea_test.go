package tui

import (
	"bytes"
	"io"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"
)

type component interface {
	Update(msg tea.Msg) tea.Cmd
}

func sendAndWait(c component, msg tea.Msg) {
	msgs := []tea.Msg{msg}
	for len(msgs) > 0 {
		msg, msgs = msgs[0], msgs[1:]
		if cmd := c.Update(msg); cmd != nil {
			msgs = append(msgs, flattenTeaMsg(cmd())...)
		}
	}
}

func flattenTeaMsg(msg tea.Msg) []tea.Msg {
	if msg == nil {
		return nil
	}
	msgs, ok := msg.(tea.BatchMsg)
	if !ok {
		return []tea.Msg{msg}
	}
	var flat []tea.Msg
	for _, cmd := range msgs {
		flat = append(flat, flattenTeaMsg(cmd())...)
	}
	return flat
}

func waitFor(t testing.TB, r io.Reader, want []byte) {
	t.Helper()
	waitForFunc(t, r, func(b []byte) bool { return bytes.Contains(b, want) })
}

func waitForFunc(t testing.TB, r io.Reader, f func([]byte) bool) {
	t.Helper()
	teatest.WaitFor(t, r, f, teatest.WithDuration(time.Second), teatest.WithCheckInterval(10*time.Millisecond))
}
