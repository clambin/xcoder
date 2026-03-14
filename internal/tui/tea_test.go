package tui

import (
	"bytes"
	"io"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/exp/teatest/v2"
)

func waitFor(t testing.TB, r io.Reader, want []byte) {
	t.Helper()
	waitForFunc(t, r, func(b []byte) bool { return bytes.Contains(b, want) })
}

func waitForFunc(t testing.TB, r io.Reader, f func([]byte) bool) {
	t.Helper()
	teatest.WaitFor(t, r, f, teatest.WithDuration(time.Second), teatest.WithCheckInterval(10*time.Millisecond))
}

type model[T any] interface {
	Init() tea.Cmd
	Update(msg tea.Msg) (T, tea.Cmd)
	View() string
}

type app[T model[T]] struct {
	model model[T]
}

func (a app[T]) Init() tea.Cmd {
	return a.model.Init()
}

func (a app[T]) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	a.model, cmd = a.model.Update(msg)
	return a, cmd
}

func (a app[T]) View() tea.View {
	return tea.NewView(a.model.View())
}
