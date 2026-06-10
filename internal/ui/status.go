package ui

import (
	"fmt"
	"strings"
	"time"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

const blinkStatusInterval = 500 * time.Millisecond

type statusLine struct {
	styles     StatusStyles
	transcoder Transcoder
	spinner    spinner.Model
	width      int
	showState  bool
}

func newStatusLine(transcoder Transcoder, styles StatusStyles, opts ...spinner.Option) statusLine {
	return statusLine{
		transcoder: transcoder,
		spinner:    spinner.New(opts...),
		styles:     styles,
	}
}

func (s statusLine) Init() tea.Cmd {
	return tea.Batch(
		func() tea.Msg { return s.spinner.Tick() },
		tea.Tick(blinkStatusInterval, func(_ time.Time) tea.Msg {
			return blinkStatusMsg{}
		}),
	)
}

func (s statusLine) Update(msg tea.Msg) (statusLine, tea.Cmd) {
	switch msg := msg.(type) {
	case spinner.TickMsg:
		var cmd tea.Cmd
		if msg.ID == s.spinner.ID() {
			s.spinner, cmd = s.spinner.Update(msg)
		}
		return s, cmd
	case blinkStatusMsg:
		s.showState = !s.showState
		return s, tea.Tick(blinkStatusInterval, func(_ time.Time) tea.Msg {
			return blinkStatusMsg{}
		})
	default:
		return s, nil
	}
}

func (s statusLine) View() string {
	state := s.state()
	return s.styles.Main.
		MaxHeight(1).
		Padding(0, 2).
		Render(lipgloss.JoinHorizontal(lipgloss.Left,
			lipgloss.NewStyle().
				Width(s.width-lipgloss.Width(state)-4).
				Padding(0, 1, 0, 0).
				Render(s.status()),
			state,
		))
}

func (s statusLine) setWidth(width int) statusLine {
	s.width = width
	return s
}

var boolToString = map[bool]string{
	true:  "ON ",
	false: "OFF",
}

func (s statusLine) state() string {
	batchState := boolToString[s.transcoder.Active()]
	if batchState == boolToString[true] {
		if s.showState {
			batchState = s.styles.Processing.Render(batchState)
		} else {
			batchState = strings.Repeat(" ", len(batchState))
		}
	}
	return lipgloss.JoinHorizontal(lipgloss.Left,
		"Overwrite target: "+boolToString[s.transcoder.OverwriteTarget()],
		" Remove source: "+boolToString[s.transcoder.RemoveSource()],
		" Batch processing: "+batchState,
	)
}

func (s statusLine) status() string {
	var status string
	if converting := s.transcoder.SessionCount(); converting > 0 {
		status = fmt.Sprintf("Converting %d file(s) ... %s", converting, s.spinner.View())
	}
	return status
}

// blinkStatusMsg is a message that blinks the state if it's "on"
type blinkStatusMsg struct{}
