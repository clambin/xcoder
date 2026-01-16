package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/clambin/xcoder/internal/pipeline"
)

type statusLine struct {
	styles    StatusStyles
	queue     Queue
	spinner   spinner.Model
	width     int
	showState bool
}

func newStatusLine(queue Queue, styles StatusStyles) *statusLine {
	return &statusLine{
		queue:   queue,
		spinner: spinner.New(spinner.WithSpinner(spinner.Dot)),
		styles:  styles,
	}
}

func (s *statusLine) Init() tea.Cmd {
	return func() tea.Msg { return s.spinner.Tick() }
}

func (s *statusLine) SetSize(width, _ int) {
	s.width = width
}

func (s *statusLine) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case spinner.TickMsg:
		var cmd tea.Cmd
		if msg.ID == s.spinner.ID() {
			s.spinner, cmd = s.spinner.Update(msg)
			s.showState = !s.showState
		}
		return cmd
	default:
		return nil
	}
}

func (s *statusLine) View() string {
	status := s.viewStatus()
	state := s.viewState()
	filler := strings.Repeat(" ", max(0, s.width-4-lipgloss.Width(status)-lipgloss.Width(state)))
	return s.styles.Main.
		Padding(0, 2, 0, 2).
		Render(
			lipgloss.JoinHorizontal(lipgloss.Left, status, filler, state),
		)
}

func (s *statusLine) viewStatus() string {
	var status string
	if converting := s.queue.Stats()[pipeline.Converting]; converting > 0 {
		status = fmt.Sprintf("Converting %d file(s) ... %s", converting, s.spinner.View())
	}
	return status
}

func (s *statusLine) viewState() string {
	var state string
	switch s.queue.Active() {
	case false:
		state = "OFF"
	case true:
		stateContent := "   "
		if s.showState {
			stateContent = "ON "
		}
		state = s.styles.Processing.Render(stateContent)
	}
	return "Batch processing: " + state
}
