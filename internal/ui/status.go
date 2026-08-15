package ui

import (
	"fmt"
	"time"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"codeberg.org/clambin/bubbles/statusbar"
)

const statusBarRefreshInterval = 500 * time.Millisecond

var boolToString = map[bool]string{
	true:  "ON",
	false: "OFF",
}

type StatusBar struct {
	statusbar        statusbar.StatusBar
	styles           StatusStyles
	transcoder       Transcoder
	profile          string
	width            int
	showProcessingOn bool
}

func newStatusBar(transcoder Transcoder, profile string, styles StatusStyles, opts ...spinner.Option) StatusBar {
	return StatusBar{
		statusbar:  statusbar.New().BaseStyle(styles.Main).Spinner(opts...),
		styles:     styles,
		transcoder: transcoder,
		profile:    profile,
	}
}

func (s StatusBar) Init() tea.Cmd {
	return tea.Batch(s.statusbar.Init(), refreshStatusBarCmd())
}

func (s StatusBar) Update(msg tea.Msg) (StatusBar, tea.Cmd) {
	switch msg := msg.(type) {
	case refreshStatusBarMsg:
		s.showProcessingOn = !s.showProcessingOn
		return s.refresh(), refreshStatusBarCmd()
	default:
		var cmd tea.Cmd
		s.statusbar, cmd = s.statusbar.Update(msg)
		return s, cmd
	}
}

func (s StatusBar) View() string {
	renderedPadding := s.styles.Main.Render("  ")
	renderedSpace := s.styles.Main.Render(" ")
	renderedConfiguration := s.viewConfiguration()
	renderedProcessingState := s.viewProcessingState()
	width := s.width - 6 - lipgloss.Width(renderedConfiguration) - lipgloss.Width(renderedProcessingState)

	return renderedPadding +
		s.statusbar.Width(width).View() +
		renderedSpace +
		renderedConfiguration +
		renderedSpace +
		renderedProcessingState +
		renderedPadding
}

func (s StatusBar) Width(width int) StatusBar {
	s.width = width
	return s.refresh()
}

func (s StatusBar) refresh() StatusBar {
	processingCount := s.transcoder.SessionCount()
	switch processingCount {
	case 0:
		s.statusbar = s.statusbar.SetStatus("")
	default:
		s.statusbar = s.statusbar.SetStatus(fmt.Sprintf("Converting %d file(s) ...", processingCount), statusbar.WithShowSpinner(true))
	}
	return s
}

func (s StatusBar) viewConfiguration() string {
	configText := "Profile: " + s.profile +
		" Overwrite target: " + boolToString[s.transcoder.OverwriteTarget()] +
		" Remove source: " + boolToString[s.transcoder.RemoveSource()]
	return s.styles.Main.Render(configText)
}

func (s StatusBar) viewProcessingState() string {
	batchProcessing := s.transcoder.Active()
	batchStateString := boolToString[batchProcessing]
	if batchProcessing {
		if !s.showProcessingOn {
			batchStateString = "   "
		} else {
			batchStateString = s.styles.Processing.Render(batchStateString + " ")
		}
	}
	return s.styles.Main.Render("Batch processing: " + batchStateString)
}

// refreshStatusBarMsg is a message that refreshes the status bar
type refreshStatusBarMsg struct{}

func refreshStatusBarCmd() tea.Cmd {
	return tea.Tick(statusBarRefreshInterval, func(_ time.Time) tea.Msg {
		return refreshStatusBarMsg{}
	})
}
