package tui

import (
	"io"
	"iter"
	"time"

	"codeberg.org/clambin/bubbles/table"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/clambin/xcoder/internal/pipeline"
)

const (
	refreshInterval = 2 * 250 * time.Millisecond
)

// activePane determines which pane is active, i.e., which one gets keyboard input and which one to display.
type activePane int

const (
	queuePane activePane = iota
	logPane
)

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Controller

// Queue is the interface for a pipeline.Queue.
type Queue interface {
	Stats() map[pipeline.Status]int
	All() iter.Seq[*pipeline.WorkItem]
	Queue(*pipeline.WorkItem)
	SetActive(bool)
	Active() bool
}

var _ tea.Model = Controller{}

// Controller implements the UI for xcoder.
type Controller struct {
	mediaFilterStyle lipgloss.Style
	queue            Queue
	statusLine       tea.Model
	configPane       configPane
	helpController   helpController
	keyMap           ControllerKeyMap
	selectedRow      table.Row
	logPane          logViewer
	queuePane        queueViewer
	filter           filter
	width            int
	height           int
}

// New returns a new Controller for the provided Queue.
func New(queue Queue, cfg pipeline.Configuration) Controller {
	styles := DefaultStyles()
	controllerKeyMap := defaultControllerKeyMap()
	filterKeyMap := defaultFilterKeyMap()

	ui := Controller{
		queue:            queue,
		configPane:       newConfigPane(cfg, styles.Config),
		statusLine:       newStatusLine(queue, styles.Status),
		queuePane:        newQueueViewer(defaultQueueViewerKeyMap(), styles.TableStyle),
		logPane:          newLogViewer(defaultLogViewerKeyMap(), styles.FrameStyle),
		filter:           filter{keyMap: filterKeyMap},
		mediaFilterStyle: styles.MediaFilter,
		keyMap:           controllerKeyMap,
		helpController:   newHelpController(controllerKeyMap, filterKeyMap, styles.Help),
	}
	return ui
}

// LogWriter returns the io.Writer for the log pane.
// Calling applications use this to direct log/slog output to the screen.
func (c Controller) LogWriter() io.Writer {
	return c.logPane.LogWriter()
}

// Init implements the tea.Model interface. It initializes the controller and all subcomponents.
func (c Controller) Init() tea.Cmd {
	return tea.Batch(
		c.statusLine.Init(),
		c.queuePane.Init(),
		c.logPane.Init(),
		setPaneCmd(queuePane),
		tea.Tick(refreshInterval, autoRefreshCmd()),
	)
}

// Update implements the tea.Model interface. It updates the controller and all subcomponents.
func (c Controller) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	// send msg to all subcomponents, except for key msgs
	if _, ok := msg.(tea.KeyMsg); !ok {
		c.helpController, cmd = c.helpController.Update(msg)
		cmds = append(cmds, cmd)
		c.filter, cmd = c.filter.Update(msg)
		cmds = append(cmds, cmd)
		c.queuePane, cmd = c.queuePane.Update(msg)
		cmds = append(cmds, cmd)
		c.logPane, cmd = c.logPane.Update(msg)
		cmds = append(cmds, cmd)
		c.statusLine, cmd = c.statusLine.Update(msg)
		cmds = append(cmds, cmd)
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// controller manages the size of the components.
		c.width = msg.Width
		c.height = msg.Height
		// how much room is left for the panes?
		paneHeight := c.height - lipgloss.Height(c.viewHeader()) - lipgloss.Height(c.viewFooter())
		cmds = append(cmds,
			c.queuePane.setSizeCmd(msg.Width, paneHeight),
			c.logPane.setSizeCmd(msg.Width, paneHeight),
		)
	case autoRefreshMsg:
		// regular refresh. issue a manual refresh and schedule the next update.
		cmds = append(cmds,
			// refresh the screen
			refreshTableCmd(),
			// set up the next refresh
			tea.Tick(refreshInterval, autoRefreshCmd()),
		)
	case filterStateChangedMsg:
		// filter state change. record the state and schedule a reload of the table.
		cmds = append(cmds, refreshTableCmd())
	case refreshTableMsg:
		// refresh the table: set the title (based on the filter) and reload the table.
		cmds = append(cmds,
			setTitleCmd(c.filter.filterState, c.mediaFilterStyle),
			loadTableCmd(c.queue.All(), c.filter.filterState, c.queuePane.showFullPath),
		)
	case table.SetRowsMsg:
		// if we don't know the selected row yet (i.e., the user hasn't scrolled yet),
		// derive it from the first table load.
		if len(msg.Rows) > 0 && c.selectedRow == nil {
			c.selectedRow = msg.Rows[0]
			// _, _ = c.LogWriter().Write([]byte("Selected row: " + c.selectedRow[0].(string) + "\n"))
		}
	case table.RowChangedMsg:
		// mark the selected row so we know which queue item to convert when the user hits <enter>.
		c.selectedRow = msg.Row
	case tea.KeyMsg:
		// if the queue pane is active and the text filter is on, it receives all keyboard inputs
		if c.queuePane.TextFilterOn() {
			c.queuePane, cmd = c.queuePane.Update(msg)
			cmds = append(cmds, cmd)
			break
		}
		switch {
		case key.Matches(msg, c.keyMap.Quit):
			// quit
			cmds = append(cmds, tea.Quit)
		case key.Matches(msg, c.keyMap.ShowLogs) && !c.logPane.active:
			// show logs (only when the log pane is not active)
			cmds = append(cmds, setPaneCmd(logPane))
		case key.Matches(msg, c.logPane.keyMap.CloseLogs) && c.logPane.active:
			// close logs (only when the log pane is active)
			cmds = append(cmds, setPaneCmd(queuePane))
		case key.Matches(msg, c.keyMap.Activate):
			// enable/disable automatic processing
			c.queue.SetActive(!c.queue.Active())
		case key.Matches(msg, c.keyMap.Convert):
			// convert selected item. we only allow this for "inspected" items
			//(i.e., not converted, rejected, skipped, etc.)
			if row := c.selectedRow; row != nil {
				workItem := row[len(row)-1].(table.UserData).Data.(*pipeline.WorkItem)
				if workItem.WorkStatus().Status == pipeline.Inspected {
					c.queue.Queue(workItem)
				}
			}
		default:
			// send unmatched key input to subcomponents.
			if c.queuePane.active {
				c.filter, cmd = c.filter.Update(msg)
				cmds = append(cmds, cmd)
			}
			c.queuePane, cmd = c.queuePane.Update(msg)
			cmds = append(cmds, cmd)
			c.logPane, cmd = c.logPane.Update(msg)
			cmds = append(cmds, cmd)
		}
	}
	return c, tea.Batch(cmds...)
}

// View implements the tea.Model interface. It renders all subcomponents.
func (c Controller) View() string {
	return lipgloss.JoinVertical(lipgloss.Top,
		c.viewHeader(),
		c.viewBody(),
		c.viewFooter(),
	)
}

func (c Controller) viewHeader() string {
	config := lipgloss.NewStyle().Padding(0, 5, 0, 0).Render(c.configPane.View())
	width := max(0, c.width-lipgloss.Width(config))
	return lipgloss.JoinHorizontal(lipgloss.Left,
		config,
		c.helpController.activeHelpPane().Width(width).View(),
	)
}

func (c Controller) viewBody() string {
	switch {
	case c.queuePane.active:
		return c.queuePane.View()
	case c.logPane.active:
		return c.logPane.View()
	default:
		return ""
	}
}

func (c Controller) viewFooter() string {
	return c.statusLine.View()
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// ControllerKeyMap

func defaultControllerKeyMap() ControllerKeyMap {
	return ControllerKeyMap{
		Quit: key.NewBinding(
			key.WithKeys(tea.KeyCtrlC.String(), "q"),
			key.WithHelp("q/ctrl+c", "quit"),
		),
		Activate: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "activate batch transcoding"),
		),
		Convert: key.NewBinding(
			key.WithKeys(tea.KeyEnter.String()),
			key.WithHelp("enter", "transcode selected item"),
		),
		FullPath: key.NewBinding(
			key.WithKeys("f"),
			key.WithHelp("f", "show full path"),
		),
		ShowLogs: key.NewBinding(
			key.WithKeys("l"),
			key.WithHelp("l", "show logs"),
		),
		CloseLogs: key.NewBinding(
			key.WithKeys(tea.KeyEscape.String(), "l"),
			key.WithHelp("esc/l", "close logs"),
		),
	}
}

func defaultFilterKeyMap() FilterKeyMap {
	return FilterKeyMap{
		ShowSkippedFiles: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "show/hide skipped files"),
		),
		ShowRejectedFiles: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "show/hide rejected files"),
		),
		ShowConvertedFiles: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "show/hide converted files")),
	}
}

var _ help.KeyMap = ControllerKeyMap{}

type ControllerKeyMap struct {
	Quit      key.Binding
	Activate  key.Binding
	Convert   key.Binding
	FullPath  key.Binding
	ShowLogs  key.Binding
	CloseLogs key.Binding
}

func (k ControllerKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.Quit,
		k.Activate,
		k.Convert,
	}
}

func (k ControllerKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		k.ShortHelp(),            // General
		{k.FullPath, k.ShowLogs}, // View
	}
}

var _ help.KeyMap = FilterKeyMap{}

type FilterKeyMap struct {
	ShowSkippedFiles   key.Binding
	ShowRejectedFiles  key.Binding
	ShowConvertedFiles key.Binding
}

func (f FilterKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		f.ShowSkippedFiles,
		f.ShowRejectedFiles,
		f.ShowConvertedFiles,
	}
}

func (f FilterKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{f.ShortHelp()}
}
