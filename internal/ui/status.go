package ui

import (
	"fmt"
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
	state := "Batch processing: " + s.state()
	status := lipgloss.NewStyle().
		Padding(0, 1, 0, 0).
		Width(max(0, s.width-lipgloss.Width(state)-4)).
		Render(s.status())
	return s.styles.Main.
		Padding(0, 2, 0, 2).
		MaxHeight(1).
		Render(lipgloss.JoinHorizontal(lipgloss.Left, status, state))
}

func (s statusLine) setWidth(width int) statusLine {
	s.width = width
	return s
}

func (s statusLine) status() string {
	var status string
	if converting := s.transcoder.SessionCount(); converting > 0 {
		status = fmt.Sprintf("Converting %d file(s) ... %s", converting, s.spinner.View())
	}
	return status
}

var stateOnContent = map[bool]string{
	true:  "ON ",
	false: "   ",
}

func (s statusLine) state() string {
	if !s.transcoder.Active() {
		return "OFF"
	}
	return s.styles.Processing.Render(stateOnContent[s.showState])
}

// blinkStatusMsg is a message that blinks the state if it's "on"
type blinkStatusMsg struct{}
