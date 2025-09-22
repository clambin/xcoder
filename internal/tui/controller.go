package tui

import (
	"os"
	"path/filepath"
	"time"

	"codeberg.org/clambin/bubbles/helper"
	"codeberg.org/clambin/bubbles/logger"
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
	helpPane         helper.Helper
	queuePane        tea.Model
	filter           filter
	width            int
	height           int
	keyMap           ControllerKeyMap
	showFullPath     bool
	selectedRow      table.Row
	logger           *logger.MessageLogger
	mediaFilterStyle lipgloss.Style
}

func New(queue Queue, cfg pipeline.Configuration) Controller {
	var l *logger.MessageLogger
	if cfg.Log.Level == "debug" {
		f, err := os.OpenFile("messages.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0664)
		if err != nil {
			panic(err)
		}
		l = logger.NewMessageLogger(f)
	}
	styles := DefaultStyles()
	ui := Controller{
		logger:     l,
		queue:      queue,
		configPane: newConfigPane(cfg, styles.Config),
		statusLine: newStatusLine(queue, styles.Status),
		keyMap:     defaultKeyMap,
		queuePane: table.NewFilterTable(
			"media files",
			columns,
			nil,
			styles.TableStyle,
			table.DefaultFilterTableKeyMap(),
		),
		filter:           filter{keyMap: defaultFilterKeyMap},
		mediaFilterStyle: styles.MediaFilter,
	}
	ui.helpPane = helper.New().Styles(styles.Help).Sections(ui.helpSections())
	return ui
}

func (c Controller) Init() tea.Cmd {
	return tea.Batch(
		c.statusLine.Init(),
		c.queuePane.Init(),
		// refreshes the UI at the specified interval
		tea.Tick(refreshInterval, func(t time.Time) tea.Msg {
			return autoRefreshMsg{}
		}),
	)
}

func (c Controller) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if c.logger != nil {
		c.logger.Log(msg)
	}
	var cmds []tea.Cmd
	var cmd tea.Cmd

	c.statusLine, cmd = c.statusLine.Update(msg)
	cmds = append(cmds, cmd)
	c.queuePane, cmd = c.queuePane.Update(msg)
	cmds = append(cmds, cmd)
	c.filter, cmd = c.filter.Update(msg)
	cmds = append(cmds, cmd)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		c.width = msg.Width
		c.height = msg.Height
		c.queuePane, cmd = c.queuePane.Update(table.SetSizeMsg{Width: msg.Width, Height: c.contentHeight()})
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
		// if we don't know the selected row yet (i.e., user hasn't scrolled yet), derive it from the first table load.
		if len(msg.Rows) > 0 && c.selectedRow == nil {
			c.selectedRow = msg.Rows[0]
		}
	case table.OnRowChangedMsg:
		c.selectedRow = msg.Row
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, c.keyMap.Quit):
			cmds = append(cmds, tea.Quit)
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
		}
	}
	return c, tea.Batch(cmds...)
}

func (c Controller) View() string {
	return lipgloss.JoinVertical(lipgloss.Top,
		c.viewHeader(),
		c.queuePane.View(),
		c.viewFooter(),
	)
}

func (c Controller) viewHeader() string {
	config := lipgloss.NewStyle().Padding(0, 5, 0, 0).Render(c.configPane.View())
	help := c.helpPane.Width(c.width - lipgloss.Width(config)).View()
	return lipgloss.JoinHorizontal(lipgloss.Left, "\n"+config, help)
}

func (c Controller) viewFooter() string {
	return c.statusLine.View()
}

func (c Controller) contentHeight() int {
	return c.height - lipgloss.Height(c.viewHeader()) - lipgloss.Height(c.viewFooter())
}

func (c Controller) helpSections() []helper.Section {
	return []helper.Section{
		{Title: "General", Keys: c.keyMap.ShortHelp()},
		{Title: "Navigation", Keys: table.DefaultKeyMap().ShortHelp()},
		{Title: "Filters", Keys: table.DefaultFilterTableKeyMap().FilterKeyMap.ShortHelp()},
		{Title: "Media filters", Keys: c.filter.keyMap.ShortHelp()},
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
