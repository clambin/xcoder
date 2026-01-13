package tui

import (
	"bytes"
	"io"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/exp/golden"
	"github.com/charmbracelet/x/exp/teatest"
	"github.com/clambin/xcoder/internal/pipeline"
	"github.com/muesli/termenv"
)

func init() {
	lipgloss.SetColorProfile(termenv.ANSI256)
}

func TestConfigView_View(t *testing.T) {
	cfg := pipeline.Configuration{
		Input:       ".",
		ProfileName: "foo",
		Active:      false,
		Remove:      true,
		Overwrite:   true,
	}
	a := newConfigView(cfg, ConfigStyles{})

	golden.RequireEqual(t, a.View())

}

func TestApplication(t *testing.T) {
	q := fakeQueue{
		queue: []*pipeline.WorkItem{
			{Source: pipeline.MediaFile{Path: "file1.mp4"}},
			{Source: pipeline.MediaFile{Path: "file2.mp4"}},
		},
	}

	cfg := pipeline.Configuration{
		Input:       ".",
		ProfileName: "foo",
		Active:      false,
		Remove:      true,
		Overwrite:   true,
	}

	var a tea.Model = New(&q, cfg)
	tm := teatest.NewTestModel(t, a, teatest.WithInitialTermSize(200, 25))
	waitFor(t, tm.Output(), []byte("waiting"))

	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
	waitFor(t, tm.Output(), []byte("converting"))

	golden.RequireEqual(t, a.View())

	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	tm.WaitFinished(t)
}

func waitFor(t testing.TB, r io.Reader, want []byte) {
	t.Helper()
	waitForFunc(t, r, func(b []byte) bool { return bytes.Contains(b, want) })
}

func waitForFunc(t testing.TB, r io.Reader, f func([]byte) bool) {
	t.Helper()
	teatest.WaitFor(t, r, f, teatest.WithDuration(time.Hour), teatest.WithCheckInterval(10*time.Millisecond))
}
