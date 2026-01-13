package tui

import (
	tea "github.com/charmbracelet/bubbletea"
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
