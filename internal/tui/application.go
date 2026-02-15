package tui

import (
	"io"
	"time"

	"codeberg.org/clambin/bubbles/helper"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/clambin/xcoder/internal/pipeline"
)

const refreshInterval = 250 * time.Millisecond

type paneID string

const (
	queuePane paneID = "queue"
	logPane   paneID = "log"
)

var _ tea.Model = Application{}

// Application is the controller for the UI.
type Application struct {
	configViewer configView
	helpViewer   helpViewer
	statusLine   tea.Model
	activePane   paneID
	panes        map[paneID]tea.Model
	keyMap       KeyMap
	width        int
	height       int
}

func New(queue Queue, config pipeline.Configuration) Application {
	styles := DefaultStyles()
	keyMap := DefaultKeyMap()

	h := map[paneID]helper.Helper{
		queuePane: helper.New().Styles(styles.Help).Sections([]helper.Section{
			{Title: "General", Keys: []key.Binding{keyMap.Quit, keyMap.ShowLogs}},
			{Title: "Navigation", Keys: keyMap.QueueViewer.FilterTableKeyMap.ShortHelp()},
			{Title: "Media", Keys: []key.Binding{keyMap.QueueViewer.ActivateQueue, keyMap.QueueViewer.Convert, keyMap.QueueViewer.ShowFullPath}},
			{Title: "Filter", Keys: keyMap.QueueViewer.MediaFilterKeyMap.ShortHelp()},
		}),
		logPane: helper.New().Styles(styles.Help).Sections([]helper.Section{
			{Title: "General", Keys: []key.Binding{keyMap.Quit}},
			{Title: "Logs", Keys: keyMap.LogViewer.ShortHelp()},
		}),
	}
	return Application{
		configViewer: newConfigView(config, styles.Config),
		helpViewer:   newHelpViewer(h, styles.Help),
		statusLine:   newStatusLine(queue, styles.Status),
		activePane:   queuePane,
		panes: map[paneID]tea.Model{
			queuePane: newQueueViewer(queue, styles.QueueViewer, keyMap.QueueViewer),
			logPane:   newLogViewer(keyMap.LogViewer, styles.LogViewer),
		},
		keyMap: keyMap,
	}
}

func (a Application) LogWriter() io.Writer {
	return a.panes[logPane].(logViewer).LogWriter()
}

func (a Application) Init() tea.Cmd {
	cmds := make([]tea.Cmd, 0, 2+len(a.panes))
	for _, p := range a.panes {
		cmds = append(cmds, p.Init())
	}
	cmds = append(cmds,
		a.statusLine.Init(),
		func() tea.Msg { return autoRefreshUIMsg{} },
	)
	return tea.Batch(cmds...)
}

func (a Application) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	//fmt.Printf("%#+v\n", msg)
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return a.resizeComponents(msg)
	case autoRefreshUIMsg:
		cmd = tea.Batch(
			func() tea.Msg { return refreshUIMsg{} },
			tea.Tick(refreshInterval, func(t time.Time) tea.Msg { return autoRefreshUIMsg{} }),
		)
	case logViewerClosedMsg:
		a.activePane = queuePane
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, a.keyMap.Quit) && !a.panes[queuePane].(queueViewer).textFilterOn:
			// hard exit: no need to process any more messages
			return a, tea.Quit
		case key.Matches(msg, a.keyMap.ShowLogs) && !a.panes[queuePane].(queueViewer).textFilterOn:
			a.activePane = logPane
		default:
			a.panes[a.activePane], cmd = a.panes[a.activePane].Update(msg)
		}
		/*
			case refreshUIMsg:
					cmds := make([]tea.Cmd, 0, 1+len(a.panes))
					var pcmd tea.Cmd
					for p := range a.panes {
						a.panes[p], pcmd = a.panes[p].Update(msg)
						cmds = append(cmds, pcmd)
					}
					a.statusLine, pcmd = a.statusLine.Update(msg)
					cmd = tea.Batch(append(cmds, pcmd)...)

		*/
	default:
		cmds := make([]tea.Cmd, 0, 1+len(a.panes))
		var pcmd tea.Cmd
		for p := range a.panes {
			a.panes[p], pcmd = a.panes[p].Update(msg)
			cmds = append(cmds, pcmd)
		}
		a.statusLine, pcmd = a.statusLine.Update(msg)
		cmd = tea.Batch(append(cmds, pcmd)...)
	}

	return a, cmd
}

func (a Application) View() string {
	return lipgloss.JoinVertical(lipgloss.Top,
		a.viewHeader(),
		a.panes[a.activePane].View(),
		a.statusLine.View(),
	)
}

func (a Application) resizeComponents(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	a.width = msg.Width
	a.height = msg.Height
	a.statusLine = a.statusLine.(statusLine).SetSize(msg.Width, 1)
	// how much room is left for the panes?
	paneHeight := a.height - lipgloss.Height(a.viewHeader()) - 1
	// update the pane sizes
	a.panes[queuePane] = a.panes[queuePane].(queueViewer).SetSize(msg.Width, paneHeight)
	a.panes[logPane] = a.panes[logPane].(logViewer).SetSize(msg.Width, paneHeight)
	return a, func() tea.Msg { return refreshUIMsg{} }
}

func (a Application) viewHeader() string {
	config := lipgloss.NewStyle().Padding(0, 5, 0, 0).Render(a.configViewer.View())
	return lipgloss.JoinHorizontal(lipgloss.Left,
		config,
		a.helpViewer.view(a.activePane, max(0, a.width-lipgloss.Width(config))),
	)
}
