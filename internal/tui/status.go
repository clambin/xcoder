package tui

import (
	"fmt"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/clambin/xcoder/internal/pipeline"
)

type statusLine struct {
	styles    StatusStyles
	queue     Queue
	spinner   spinner.Model
	width     int
	showState bool
}

func newStatusLine(queue Queue, styles StatusStyles, opts ...spinner.Option) statusLine {
	return statusLine{
		queue:   queue,
		spinner: spinner.New(opts...),
		styles:  styles,
	}
}

func (s statusLine) Init() tea.Cmd {
	return func() tea.Msg { return s.spinner.Tick() }
}

func (s statusLine) SetSize(width, _ int) statusLine {
	s.width = width
	return s
}

func (s statusLine) Update(msg tea.Msg) (statusLine, tea.Cmd) {
	switch msg := msg.(type) {
	case spinner.TickMsg:
		var cmd tea.Cmd
		if msg.ID == s.spinner.ID() {
			s.spinner, cmd = s.spinner.Update(msg)
			s.showState = !s.showState
		}
		return s, cmd
	default:
		return s, nil
	}
}

func (s statusLine) View() string {
	state := "Batch processing: " + s.state()
	statusStyle := lipgloss.NewStyle().
		Padding(0, 1, 0, 0).
		Width(max(0, s.width-lipgloss.Width(state)-4))
	return s.styles.Main.
		Padding(0, 2, 0, 2).
		MaxHeight(1).
		Render(
			lipgloss.JoinHorizontal(lipgloss.Left,
				statusStyle.Render(s.status()),
				state,
			),
		)
}

func (s statusLine) status() string {
	var status string
	if converting := s.queue.Stats()[pipeline.Converting]; converting > 0 {
		status = fmt.Sprintf("Converting %d file(s) ... %s", converting, s.spinner.View())
	}
	return status
}

var stateOnContent = map[bool]string{
	true:  "ON ",
	false: "   ",
}

func (s statusLine) state() string {
	if !s.queue.Active() {
		return "OFF"
	}
	return s.styles.Processing.Render(stateOnContent[s.showState])
}
