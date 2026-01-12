package tui

import (
	"io"
	"iter"
	"time"

	"codeberg.org/clambin/bubbles/table"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/clambin/xcoder/internal/pipeline"
	"github.com/clambin/xcoder/internal/tui/pane"
)

const (
	refreshInterval           = 2 * 250 * time.Millisecond
	queuePane       pane.Name = "queue"
	logPane         pane.Name = "log"
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
	mediaFilter      tea.Model
	panes            tea.Model
	logWriter        io.Writer
	helpController   helpController
	configPane       configPane
	keyMap           KeyMap
	selectedRow      table.Row
	width            int
	height           int
	mediaFilterState mediaFilterState
	showFullPath     bool
	textFilterOn     bool
}

// New returns a new Controller for the provided Queue.
func New(queue Queue, cfg pipeline.Configuration) Controller {
	styles := DefaultStyles()
	keyMap := defaultKeyMap()

	lv := newLogViewer(keyMap.LogViewer, styles.FrameStyle)
	qv := newQueueViewer(keyMap.QueueViewer, styles.TableStyle)

	ui := Controller{
		queue:            queue,
		configPane:       newConfigPane(cfg, styles.Config),
		statusLine:       newStatusLine(queue, styles.Status),
		mediaFilter:      mediaFilter{keyMap: keyMap.Filter},
		mediaFilterStyle: styles.MediaFilter,
		keyMap:           keyMap,
		helpController:   newHelpController(keyMap, styles.Help),
		panes: pane.New(map[pane.Name]tea.Model{
			queuePane: qv,
			logPane:   lv,
		}),
		logWriter: lv.LogWriter(),
	}
	return ui
}

// LogWriter returns the io.Writer for the log pane.
// Calling applications use this to direct log/slog output to the screen.
func (c Controller) LogWriter() io.Writer {
	return c.logWriter
}

// Init implements the tea.Model interface. It initializes the controller and all subcomponents.
func (c Controller) Init() tea.Cmd {
	return tea.Batch(
		c.statusLine.Init(),
		c.panes.Init(),
		c.mediaFilter.Init(),
		mediaFilterActivateCmd(true),
		cmd(pane.ActivateMsg{Pane: queuePane}),
		tea.Tick(refreshInterval, autoRefreshCmd()),
	)
}

// Update implements the tea.Model interface. It updates the controller and all subcomponents.
func (c Controller) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	// send msg to all subcomponents, except for key messages
	if _, ok := msg.(tea.KeyMsg); !ok {
		c.helpController, cmd = c.helpController.Update(msg)
		cmds = append(cmds, cmd)
		c.mediaFilter, cmd = c.mediaFilter.Update(msg)
		cmds = append(cmds, cmd)
		c.panes, cmd = c.panes.Update(msg)
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
		cmds = append(cmds, paneSizeCmd(msg.Width, paneHeight))
	case showLogsMsg:
		paneNames := map[bool]pane.Name{true: logPane, false: queuePane}
		c.mediaFilter, _ = c.mediaFilter.Update(mediaFilterActivateMsg{active: !msg.on})
		c.panes, cmd = c.panes.Update(pane.ActivateMsg{Pane: paneNames[msg.on]})
		c.helpController, cmd = c.helpController.Update(pane.ActivateMsg{Pane: paneNames[msg.on]})
		cmds = append(cmds, cmd)
	case showFullPathMsg:
		c.showFullPath = msg.on
		cmds = append(cmds, refreshTableCmd())
	case textFilterStateChangMsg:
		c.textFilterOn = msg.on
		cmds = append(cmds, refreshTableCmd())
	case autoRefreshMsg:
		// regular refresh. issue a manual refresh and schedule the next update.
		cmds = append(cmds,
			// refresh the screen
			refreshTableCmd(),
			// set up the next refresh
			tea.Tick(refreshInterval, autoRefreshCmd()),
		)
	case mediaFilterChangedMsg:
		// mediaFilter state change. record the state and schedule a reload of the table.
		c.mediaFilterState = mediaFilterState(msg)
		cmds = append(cmds, refreshTableCmd())
	case refreshTableMsg:
		// refresh the table: set the title (based active the mediaFilter) and reload the table.
		// TODO: this can just be a loadTableCmd(): title can get refreshed by MediaFilterChangedMsg
		cmds = append(cmds,
			setTitleCmd(c.mediaFilterState, c.mediaFilterStyle),
			loadTableCmd(c.queue.All(), c.mediaFilterState, c.showFullPath),
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
		c, cmd = c.handleKeyMsg(msg)
		cmds = append(cmds, cmd)
	}
	return c, tea.Batch(cmds...)
}

// handleKeyMsg handles key inputs.
func (c Controller) handleKeyMsg(msg tea.KeyMsg) (Controller, tea.Cmd) {
	var cmd tea.Cmd
	// if the queue pane is active and the text filter is active, it receives all keyboard inputs
	if c.textFilterOn {
		c.panes, cmd = c.panes.Update(msg)
		return c, cmd
	}
	switch {
	case key.Matches(msg, c.keyMap.Controller.Quit):
		// quit
		return c, tea.Quit
	case key.Matches(msg, c.keyMap.Controller.ShowLogs):
		// TODO: this also catches 'l' being pressed when the log pane is active.
		// In there, it should close the logs pane, not open it again.
		//
		// Should we move all queueViewer commands to its own keyMap, help and Update handler?

		// show logs (only when the log pane is not active)
		return c, showLogsCmd(true)
	case key.Matches(msg, c.keyMap.Controller.Activate):
		// enable/disable automatic processing
		c.queue.SetActive(!c.queue.Active())
		return c, nil
	case key.Matches(msg, c.keyMap.Controller.Convert):
		// convert selected item. we only allow this for "inspected" items
		//(i.e., not converted, rejected, skipped, etc.)
		if row := c.selectedRow; row != nil {
			workItem := row[len(row)-1].(table.UserData).Data.(*pipeline.WorkItem)
			if workItem.WorkStatus().Status == pipeline.Inspected {
				c.queue.Queue(workItem)
			}
		}
		return c, nil
	default:
		var cmds []tea.Cmd
		c.mediaFilter, cmd = c.mediaFilter.Update(msg)
		cmds = append(cmds, cmd)
		c.panes, cmd = c.panes.Update(msg)
		cmds = append(cmds, cmd)
		return c, tea.Batch(cmds...)
	}
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
	return c.panes.View()
}

func (c Controller) viewFooter() string {
	return c.statusLine.View()
}
