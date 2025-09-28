package tui

import (
	"io"
	"path/filepath"
	"time"

	"codeberg.org/clambin/bubbles/helper"
	"codeberg.org/clambin/bubbles/stream"
	"codeberg.org/clambin/bubbles/table"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/clambin/xcoder/internal/pipeline"
)

const (
	refreshInterval = 2 * 250 * time.Millisecond
)

var (
	columns = []table.Column{
		{Name: "SOURCE"},
		{Name: "SOURCE STATS", Width: 25},
		{Name: "TARGET STATS", Width: 25},
		{Name: "STATUS", Width: 10, RowStyle: table.CellStyle{Style: lipgloss.NewStyle().Transform(table.StringStyler(statusColors))}},
		{Name: "COMPLETED", Width: 10, RowStyle: table.CellStyle{Style: lipgloss.NewStyle().Align(lipgloss.Right)}},
		{Name: "REMAINING", Width: 10, RowStyle: table.CellStyle{Style: lipgloss.NewStyle().Align(lipgloss.Right)}},
		{Name: "ERROR", Width: 40},
	}
)

// autoRefreshMsg refreshes the screen at a regular interval.
type autoRefreshMsg struct{}

// refreshMsg manually refreshes the screen.
// This is used to perform a manual refresh of the table (e.g., after a filter change)
type refreshMsg struct{}

type activePane int

const (
	queuePane activePane = iota
	logPane
)

type Queue interface {
	Stats() map[pipeline.Status]int
	List() []*pipeline.WorkItem
	Queue(*pipeline.WorkItem)
	SetActive(bool)
	Active() bool
}

var _ tea.Model = Controller{}

type Controller struct {
	queue            Queue
	configPane       configPane
	statusLine       tea.Model
	tableHelp        helper.Helper
	logHelp          helper.Helper
	queuePane        tea.Model
	logPane          tea.Model
	filter           filter
	width            int
	height           int
	keyMap           ControllerKeyMap
	showFullPath     bool
	selectedRow      table.Row
	mediaFilterStyle lipgloss.Style
	activePane       activePane
	textFilterOn     bool
}

func New(queue Queue, cfg pipeline.Configuration) Controller {
	styles := DefaultStyles()
	ui := Controller{
		queue:      queue,
		configPane: newConfigPane(cfg, styles.Config),
		statusLine: newStatusLine(queue, styles.Status),
		queuePane: table.NewFilterTable(
			"media files",
			columns,
			nil,
			styles.TableStyle,
			table.DefaultFilterTableKeyMap(),
		),
		logPane: logViewer{
			Model:       stream.NewStream(80, 25, stream.WithShowToggles(true)),
			frameStyles: styles.FrameStyle,
		},
		filter:           filter{keyMap: defaultFilterKeyMap},
		mediaFilterStyle: styles.MediaFilter,
		keyMap:           defaultKeyMap,
	}
	ui.tableHelp = helper.New().Styles(styles.Help).Sections(ui.tableHelpSections())
	ui.logHelp = helper.New().Styles(styles.Help).Sections(ui.logHelpSections())
	return ui
}

func (c Controller) LogWriter() io.Writer {
	return c.logPane.(logViewer).Model.(io.Writer)
}

func (c Controller) Init() tea.Cmd {
	return tea.Batch(
		c.statusLine.Init(),
		c.queuePane.Init(),
		c.logPane.Init(),
		tea.Tick(refreshInterval, func(t time.Time) tea.Msg {
			return autoRefreshMsg{}
		}),
	)
}

func (c Controller) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	// process control messages first
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		c.width = msg.Width
		c.height = msg.Height
		c.tableHelp = c.tableHelp.Width(msg.Width - lipgloss.Width(c.configPane.View()))
		c.logHelp = c.logHelp.Width(msg.Width - lipgloss.Width(c.configPane.View()))
		c.queuePane, cmd = c.queuePane.Update(table.SetSizeMsg{Width: msg.Width, Height: c.contentHeight()})
		cmds = append(cmds, cmd)
		c.logPane, cmd = c.logPane.Update(stream.SetSizeMsg{Width: msg.Width, Height: c.contentHeight()})
		cmds = append(cmds, cmd)
		c.statusLine, cmd = c.statusLine.Update(msg)
		cmds = append(cmds, cmd)
	case autoRefreshMsg:
		cmds = append(cmds,
			// refresh the screen
			func() tea.Msg { return refreshMsg{} },
			// set up the next refresh
			tea.Tick(refreshInterval, func(t time.Time) tea.Msg {
				return autoRefreshMsg{}
			}),
		)
	case refreshMsg:
		cmds = append(cmds,
			// set the title according to the filter
			c.setTitleCmd(),
			// refresh the table
			c.refreshTableCmd(),
		)
	case table.SetRowsMsg:
		// if we don't know the selected row yet (i.e., the user hasn't scrolled yet),
		// derive it from the first table load.
		if len(msg.Rows) > 0 && c.selectedRow == nil {
			c.selectedRow = msg.Rows[0]
		}
		c.queuePane, cmd = c.queuePane.Update(table.SetRowsMsg{Rows: msg.Rows})
		cmds = append(cmds, cmd)
	case table.RowChangedMsg:
		c.selectedRow = msg.Row
	case table.FilterStateChangeMsg:
		c.textFilterOn = msg.State
	case tea.KeyMsg:
		if key.Matches(msg, c.keyMap.Quit) {
			return c, tea.Quit
		}
		switch c.activePane {
		case logPane:
			switch {
			case key.Matches(msg, c.keyMap.CloseLogs):
				c.activePane = queuePane
			default:
				c.logPane, cmd = c.logPane.Update(msg)
				cmds = append(cmds, cmd)
			}
		case queuePane:
			// if the text filter is on, it receives all inputs.
			if c.textFilterOn {
				var cmd tea.Cmd
				c.queuePane, cmd = c.queuePane.Update(msg)
				return c, cmd
			}
			switch {
			case key.Matches(msg, c.keyMap.FullPath):
				c.showFullPath = !c.showFullPath
				cmds = append(cmds, c.refreshTableCmd())
			case key.Matches(msg, c.keyMap.Activate):
				c.queue.SetActive(!c.queue.Active())
			case key.Matches(msg, c.keyMap.Convert):
				if row := c.selectedRow; row != nil && row[3] == "inspected" {
					workItem := row[len(row)-1].(table.UserData).Data.(*pipeline.WorkItem)
					c.queue.Queue(workItem)
				}
			case key.Matches(msg, c.keyMap.ShowLogs):
				c.activePane = logPane
			default:
				c.filter, cmd = c.filter.Update(msg)
				cmds = append(cmds, cmd)
				c.queuePane, cmd = c.queuePane.Update(msg)
				cmds = append(cmds, cmd)
			}
		}
	default:
		c.statusLine, cmd = c.statusLine.Update(msg)
		cmds = append(cmds, cmd)
		c.filter, cmd = c.filter.Update(msg)
		cmds = append(cmds, cmd)
		c.queuePane, cmd = c.queuePane.Update(msg)
		cmds = append(cmds, cmd)
		c.logPane, cmd = c.logPane.Update(msg)
		cmds = append(cmds, cmd)
	}
	return c, tea.Batch(cmds...)
}

func (c Controller) View() string {
	return lipgloss.JoinVertical(lipgloss.Top,
		c.viewHeader(),
		c.viewBody(),
		c.viewFooter(),
	)
}

func (c Controller) viewHeader() string {
	config := lipgloss.NewStyle().Padding(0, 5, 0, 0).Render(c.configPane.View())
	var help string
	switch c.activePane {
	case queuePane:
		help = c.tableHelp.View()
	case logPane:
		help = c.logHelp.View()
	}
	return lipgloss.JoinHorizontal(lipgloss.Left, config, help)
}

func (c Controller) viewBody() string {
	switch c.activePane {
	case logPane:
		return c.logPane.View()
	case queuePane:
		return c.queuePane.View()
	default:
		return ""
	}
}

func (c Controller) viewFooter() string {
	return c.statusLine.View()
}

func (c Controller) contentHeight() int {
	return c.height - lipgloss.Height(c.viewHeader()) - lipgloss.Height(c.viewFooter())
}

func (c Controller) tableHelpSections() []helper.Section {
	h := c.keyMap.FullHelp()
	filterBindings := c.filter.keyMap.ShortHelp()
	filterBindings = append(filterBindings, table.DefaultFilterTableKeyMap().FilterKeyMap.ShortHelp()...)
	return []helper.Section{
		{Title: "General", Keys: h[0]},
		{Title: "View", Keys: h[1]},
		{Title: "Navigation", Keys: table.DefaultKeyMap().ShortHelp()},
		{Title: "Filters", Keys: filterBindings},
	}
}

func (c Controller) logHelpSections() []helper.Section {
	return []helper.Section{
		{Title: "General", Keys: c.keyMap.ShortHelp()},
		{Title: "Navigation", Keys: []key.Binding{c.keyMap.CloseLogs}},
	}
}

// setTitleCmd returns a command that sets the title of the table's frame.
func (c Controller) setTitleCmd() tea.Cmd {
	return func() tea.Msg {
		args := c.filter.String()
		if args != "" {
			args = "(" + c.mediaFilterStyle.Render(args) + ")"
		}
		return table.SetTitleMsg{Title: "media files" + args}
	}
}

// refreshTableCmd returns a command that refreshes the table with the current queue state.
func (c Controller) refreshTableCmd() tea.Cmd {
	return func() tea.Msg {
		items := c.queue.List()
		rows := make([]table.Row, 0, len(items))
		for _, item := range items {
			if !c.filter.Show(item) {
				continue
			}
			source := item.Source.Path
			if !c.showFullPath {
				source = filepath.Base(source)
			}
			workStatus := item.WorkStatus()
			var errString string
			if workStatus.Err != nil {
				errString = workStatus.Err.Error()
			}
			rows = append(rows, table.Row{
				source,
				item.SourceVideoStats().String(),
				item.TargetVideoStats().String(),
				workStatus.Status.String(),
				item.CompletedFormatted(),
				item.RemainingFormatted(),
				errString,
				table.UserData{Data: item},
			})
		}
		return table.SetRowsMsg{Rows: rows}
	}
}
