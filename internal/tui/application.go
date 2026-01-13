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
	// TODO: map[paneID]component?
	configViewer configView
	helpViewer   helpViewer
	queueViewer  *QueueViewer
	logViewer    *LogViewer
	statusLine   *statusLine
	activePane   paneID
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
			{Title: "General", Keys: []key.Binding{keyMap.Quit, keyMap.ShowLogs}},
			{Title: "Logs", Keys: keyMap.LogViewer.ShortHelp()},
		}),
	}
	return Application{
		configViewer: newConfigView(config, styles.Config),
		queueViewer:  NewQueueViewer(queue, styles.QueueViewer, keyMap.QueueViewer),
		logViewer:    NewLogViewer(keyMap.LogViewer, styles.LogViewer),
		helpViewer:   newHelpViewer(h, styles.Help),
		statusLine:   newStatusLine(queue, styles.Status),
		activePane:   queuePane,
		keyMap:       keyMap,
	}
}

func (a Application) LogWriter() io.Writer {
	return a.logViewer.LogWriter()
}

func (a Application) Init() tea.Cmd {
	return tea.Batch(
		a.queueViewer.Init(),
		a.logViewer.Init(),
		tea.Tick(refreshInterval, func(t time.Time) tea.Msg {
			return AutoRefreshUIMsg{}
		}),
	)
}

func (a Application) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	//fmt.Printf("%#+v\n", msg)
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return a.resizeComponents(msg)
	case AutoRefreshUIMsg:
		cmd = tea.Batch(
			func() tea.Msg { return RefreshUIMsg{} },
			tea.Tick(refreshInterval, func(t time.Time) tea.Msg { return AutoRefreshUIMsg{} }),
		)
	case RefreshUIMsg:
		cmd = tea.Batch(
			a.queueViewer.Update(msg),
			a.logViewer.Update(msg),
			a.statusLine.Update(msg),
		)
	case LogViewerClosedMsg:
		a.activePane = queuePane
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, a.keyMap.Quit):
			// hard exit: no need to process any more messages
			return a, tea.Quit
		case key.Matches(msg, a.keyMap.ShowLogs):
			a.activePane = logPane
		default:
			// TODO: this means only queueViewer and logViewer can handle key events.
			// Do any others need it?  MediaFilter is handled by queueViewer. Anyone else?
			switch a.activePane {
			case queuePane:
				cmd = a.queueViewer.Update(msg)
			case logPane:
				cmd = a.logViewer.Update(msg)
			}
		}
	default:
		cmd = tea.Batch(
			a.queueViewer.Update(msg),
			a.logViewer.Update(msg),
			a.statusLine.Update(msg),
		)
	}

	return a, cmd
}

func (a Application) View() string {
	return lipgloss.JoinVertical(lipgloss.Top,
		a.viewHeader(),
		a.viewBody(),
		a.viewFooter(),
	)
}

func (a Application) resizeComponents(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	a.width = msg.Width
	a.height = msg.Height
	a.statusLine.SetSize(msg.Width, 1)
	// how much room is left for the panes?
	paneHeight := a.height - lipgloss.Height(a.viewHeader()) - lipgloss.Height(a.viewFooter())
	// update the pane sizes
	a.queueViewer.SetSize(msg.Width, paneHeight)
	a.logViewer.SetSize(msg.Width, paneHeight)
	return a, func() tea.Msg { return RefreshUIMsg{} }
}

func (a Application) viewHeader() string {
	config := lipgloss.NewStyle().Padding(0, 5, 0, 0).Render(a.configViewer.View())
	width := max(0, a.width-lipgloss.Width(config))
	return lipgloss.JoinHorizontal(lipgloss.Left,
		config,
		a.helpViewer.view(a.activePane, width),
	)
}

func (a Application) viewBody() string {
	switch a.activePane {
	case queuePane:
		return a.queueViewer.View()
	case logPane:
		return a.logViewer.View()
	default:
		return ""
	}
}

func (a Application) viewFooter() string {
	return a.statusLine.View()
}
