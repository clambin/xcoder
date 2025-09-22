package tui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/clambin/xcoder/internal/pipeline"
)

const (
	stateBlinkInterval = 500 * time.Millisecond
)

var _ tea.Model = statusLine{}

type statusLine struct {
	queue     Queue
	spinner   spinner.Model
	width     int
	showState bool
	styles    StatusStyles
}

func newStatusLine(queue Queue, styles StatusStyles) statusLine {
	s := spinner.New(spinner.WithSpinner(spinner.Dot))
	//s.Spinner.FPS = 250 * time.Millisecond
	return statusLine{
		queue:   queue,
		spinner: s,
		styles:  styles,
	}
}

func (s statusLine) Init() tea.Cmd {
	return tea.Batch(
		func() tea.Msg { return s.spinner.Tick() },
		tea.Tick(stateBlinkInterval, func(t time.Time) tea.Msg { return blinkStateMsg{} }),
	)
}

func (s statusLine) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.width = msg.Width
		return s, nil
	case spinner.TickMsg:
		var cmd tea.Cmd
		s.spinner, cmd = s.spinner.Update(msg)
		return s, cmd
	case blinkStateMsg:
		s.showState = !s.showState
		return s, tea.Tick(stateBlinkInterval, func(t time.Time) tea.Msg { return blinkStateMsg{} })
	}
	return s, nil
}

func (s statusLine) View() string {
	status := s.viewStatus()
	state := s.viewState()
	filler := s.styles.Main.Width(s.width - lipgloss.Width(status) - lipgloss.Width(state)).Render(" ")
	return lipgloss.JoinHorizontal(lipgloss.Left, status, filler, state)
}

func (s statusLine) viewStatus() string {
	var status string
	if converting := s.queue.Stats()[pipeline.Converting]; converting > 0 {
		status = fmt.Sprintf("Converting %d file(s) ... %s", converting, s.spinner.View())
	}
	return s.styles.Main.Padding(0, 0, 0, 2).Render(status)
}

func (s statusLine) viewState() string {
	var state string
	switch s.queue.Active() {
	case false:
		state = "OFF"
	case true:
		stateContent := "  "
		if s.showState {
			stateContent = "ON"
		}
		state = s.styles.Processing.Render(stateContent)
	}
	return s.styles.Main.Padding(0, 2, 0, 0).Render("Batch processing: " + state)
}

type blinkStateMsg struct{}
