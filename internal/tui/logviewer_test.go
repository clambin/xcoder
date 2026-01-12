package tui

/*
import (
	"bytes"
	"testing"
	"time"

	"codeberg.org/clambin/bubbles/stream"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"
)

func TestLogViewer(t *testing.T) {
	l := newLogViewer(
		defaultLogViewerKeyMap(),
		DefaultStyles().FrameStyle,
	)
	_, _ = l.LogWriter().Write([]byte("Hello World\nNice to see you\n"))
	tm := teatest.NewTestModel(t, logViewerModel{l})
	tm.Send(stream.SetSizeMsg{Width: 40, Height: 6})            // set the size of the log pane
	tm.Send(setPaneMsg(logPane))                                // activate the log pane
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'w'}}) // switch on word wrap

	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return bytes.Contains(bts, []byte("Nice to see you"))
	}, teatest.WithDuration(time.Second), teatest.WithCheckInterval(10*time.Millisecond),
	)
}

var (
	_ tea.Model = logViewerModel{}
)

type logViewerModel struct {
	logViewer
}

func (m logViewerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.logViewer, cmd = m.logViewer.Update(msg)
	return m, cmd
}
*/
