package refactor

import (
	"fmt"
	"iter"
	"path/filepath"
	"time"

	"codeberg.org/clambin/bubbles/colors"
	"codeberg.org/clambin/bubbles/table"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/clambin/xcoder/internal/pipeline"
)

const refreshInterval = 500 * time.Millisecond

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

	statusColors = map[string]lipgloss.Style{
		pipeline.Rejected.String():   lipgloss.NewStyle().Foreground(colors.IndianRed),
		pipeline.Skipped.String():    lipgloss.NewStyle().Foreground(colors.Green),
		pipeline.Converted.String():  lipgloss.NewStyle().Foreground(colors.Yellow4Alt),
		pipeline.Converting.String(): lipgloss.NewStyle().Foreground(colors.Orange1),
		pipeline.Failed.String():     lipgloss.NewStyle().Foreground(colors.Red),
	}
)

// Queue is the interface for a pipeline.Queue.
type Queue interface {
	Queue(item *pipeline.WorkItem)
	SetActive(active bool)
	All() iter.Seq[*pipeline.WorkItem]
	Active() bool
}

var _ Queue = (*pipeline.Queue)(nil)

type QueueViewer struct {
	table           tea.Model
	queue           Queue
	keyMap          QueueViewerKeyMap
	selectedRow     table.Row
	mediaFilter     MediaFilter
	refreshInterval time.Duration
	textFilterOn    bool
	showFullPath    bool
}

func NewQueueViewer(queue Queue, styles table.FilterTableStyles, keyMap QueueViewerKeyMap) *QueueViewer {
	return &QueueViewer{
		mediaFilter:     MediaFilter{KeyMap: keyMap.MediaFilterKeyMap},
		table:           table.NewFilterTable("media files", columns, nil, styles, keyMap.FilterTableKeyMap),
		keyMap:          keyMap,
		queue:           queue,
		refreshInterval: refreshInterval,
	}
}

func (m *QueueViewer) Init() tea.Cmd {
	// start queue update ticker
	// TODO: move this to the Controller
	//return tea.Batch(
	//	func() tea.Msg { return refreshQueueMsg{} },
	//	tea.Tick(m.refreshInterval, autoRefreshQueueCmd()),
	//)
	return nil
}

func (m *QueueViewer) SetSize(width, height int) {
	var cmd tea.Cmd
	// TODO: table doesn't lend itself well to SetSize() approach. Should have its own SetSize method
	// TODO: QueueViewer should draw its frame (+title) itself
	m.table, cmd = m.table.Update(table.SetSizeMsg{Width: width, Height: height})
	for cmd != nil {
		m.table, cmd = m.table.Update(cmd())
	}
}

func (m *QueueViewer) Update(msg tea.Msg) tea.Cmd {
	fmt.Printf("%#+v\n", msg)
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case table.FilterStateChangeMsg:
		m.textFilterOn = msg.State
	case MediaFilterChangedMsg:
		cmd = func() tea.Msg { return RefreshUIMsg{} }
	case RefreshUIMsg:
		cmd = loadTableCmd(m.queue.All(), m.mediaFilter.mediaFilterState, m.showFullPath)
	case table.RowChangedMsg:
		// mark the selected row so we know which queue item to convert when the user hits <enter>.
		// TODO: does this consistently work? even if table is first loaded?
		m.selectedRow = msg.Row
	case tea.KeyMsg:
		// if the text filter is active, it receives all inputs.
		if m.textFilterOn {
			m.table, cmd = m.table.Update(msg)
			break
		}
		switch {
		case key.Matches(msg, m.keyMap.ActivateQueue):
			// toggle queue active state
			m.queue.SetActive(!m.queue.Active())
		case key.Matches(msg, m.keyMap.Convert):
			if row := m.selectedRow; row != nil {
				m.queue.Queue(row[len(row)-1].(table.UserData).Data.(*pipeline.WorkItem))
			}
		case key.Matches(msg, m.keyMap.ShowFullPath):
			m.showFullPath = !m.showFullPath
			// refresh the table
			cmd = func() tea.Msg { return RefreshUIMsg{} }
		default:
			// route key to mediaFilter
			if cmd = m.mediaFilter.Update(msg); cmd == nil {
				// if no action, route to table
				m.table, cmd = m.table.Update(msg)
			}
		}
	default:
		// any other message is passed to the table
		m.table, cmd = m.table.Update(msg)
	}
	return cmd
}

func (m *QueueViewer) View() string {
	return m.table.View()
}

// loadTableCmd builds the table with the current Queue state and issues a command to load it in the table.
func loadTableCmd(items iter.Seq[*pipeline.WorkItem], f MediaFilterState, showFullPath bool) tea.Cmd {
	return func() tea.Msg {
		var rows []table.Row
		for item := range items {
			if !f.Show(item) {
				continue
			}
			source := item.Source.Path
			if !showFullPath {
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
