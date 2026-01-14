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
			msg := cmd()
			switch msg.(type) {
			case tea.BatchMsg:
				for _, c := range msg.(tea.BatchMsg) {
					if m := c(); m != nil {
						msgs = append(msgs, m)
					}
				}
			default:
				msgs = append(msgs, msg)
			}
		}
	}
}

type msgQueue []tea.Msg

func (q *msgQueue) push(msg tea.Msg) {
	if batch, ok := msg.(tea.BatchMsg); ok {
		for _, m := range batch {
			q.push(m())
		}
	} else {
		*q = append(*q, msg)
	}
}

func (q *msgQueue) pop() tea.Msg {
	if len(*q) == 0 {
		return nil
	}
	var msg tea.Msg
	msg, *q = (*q)[0], (*q)[1:]
	return msg
}

func waitFor(t testing.TB, r io.Reader, want []byte) {
	t.Helper()
	waitForFunc(t, r, func(b []byte) bool { return bytes.Contains(b, want) })
}

func waitForFunc(t testing.TB, r io.Reader, f func([]byte) bool) {
	t.Helper()
	teatest.WaitFor(t, r, f, teatest.WithDuration(time.Second), teatest.WithCheckInterval(10*time.Millisecond))
}
