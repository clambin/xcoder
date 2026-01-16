package tui

import (
	"path/filepath"

	"codeberg.org/clambin/bubbles/table"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/clambin/xcoder/internal/pipeline"
)

const title = "media files"

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

// Queue is the interface for a pipeline.Queue.
type Queue interface {
	Queue(item *pipeline.WorkItem)
	SetActive(active bool)
	All() []*pipeline.WorkItem
	Active() bool
	Stats() map[pipeline.Status]int
}

var _ Queue = (*pipeline.Queue)(nil)

type queueViewer struct {
	mediaFilterStyle lipgloss.Style
	table            *table.FilterTable
	queue            Queue
	keyMap           QueueViewerKeyMap
	mediaFilter      mediaFilter
	textFilterOn     bool
	showFullPath     bool
}

func newQueueViewer(queue Queue, styles QueueViewerStyles, keyMap QueueViewerKeyMap) *queueViewer {
	return &queueViewer{
		table:            table.NewFilterTable(title, columns, nil, styles.Table, keyMap.FilterTableKeyMap),
		queue:            queue,
		keyMap:           keyMap,
		mediaFilter:      mediaFilter{KeyMap: keyMap.MediaFilterKeyMap},
		mediaFilterStyle: styles.MediaFilter,
	}
}

func (q *queueViewer) Init() tea.Cmd {
	return nil
}

func (q *queueViewer) SetSize(width, height int) {
	q.table.SetSize(width, height)
}

func (q *queueViewer) Update(msg tea.Msg) tea.Cmd {
	// fmt.Printf("msg: %#+v\n", msg)

	var cmd tea.Cmd
	switch msg := msg.(type) {
	case table.FilterStateChangeMsg:
		// text filter changed.  if on, we need to route all keys to the table.
		q.textFilterOn = msg.State
	case refreshUIMsg:
		// refresh the table. this is done in a cmd, so it doesn't block the UI loop.
		cmd = loadTableCmd(q.queue.All(), q.mediaFilter.mediaFilterState, q.showFullPath)
	case mediaFilterChangedMsg:
		// filter changed. change the table title
		newTitle := title
		if filter := q.mediaFilter.mediaFilterState.String(); filter != "" {
			newTitle += " (" + q.mediaFilterStyle.Render(filter) + ")"
		}
		q.table.SetTitle(newTitle)
	case setRowsMsg:
		q.table.SetRows(msg.rows)
	case tea.KeyMsg:
		// if the text filter is active, it receives all inputs.
		if q.textFilterOn {
			cmd = q.table.Update(msg)
			break
		}
		switch {
		case key.Matches(msg, q.keyMap.ActivateQueue):
			// toggle queue active state
			q.queue.SetActive(!q.queue.Active())
		case key.Matches(msg, q.keyMap.Convert):
			if row := q.table.SelectedRow; row != nil {
				q.queue.Queue(row[len(row)-1].(table.UserData).Data.(*pipeline.WorkItem))
				cmd = func() tea.Msg { return refreshUIMsg{} }
			}
		case key.Matches(msg, q.keyMap.ShowFullPath):
			q.showFullPath = !q.showFullPath
			// refresh the table
			cmd = func() tea.Msg { return refreshUIMsg{} }
		default:
			// route key to mediaFilter
			if cmd = q.mediaFilter.Update(msg); cmd == nil {
				// if no action, route to table
				cmd = q.table.Update(msg)
			}
		}
	default:
		// any other message is passed to the table
		cmd = q.table.Update(msg)
	}
	return cmd
}

func (q *queueViewer) View() string {
	return q.table.View()
}

// loadTableCmd builds the table with the current Queue state and issues a command to load it in the table.
func loadTableCmd(items []*pipeline.WorkItem, f mediaFilterState, showFullPath bool) tea.Cmd {
	return func() tea.Msg {
		var rows []table.Row
		for _, item := range items {
			if f.Show(item) {
				rows = append(rows, itemToRow(item, showFullPath))
			}
		}
		return setRowsMsg{rows: rows}
	}
}

func itemToRow(item *pipeline.WorkItem, showFullPath bool) table.Row {
	source := item.Source.Path
	if !showFullPath {
		source = filepath.Base(source)
	}
	workStatus := item.WorkStatus()
	var errString string
	if workStatus.Err != nil {
		errString = workStatus.Err.Error()
	}

	return table.Row{
		source,
		item.SourceVideoStats().String(),
		item.TargetVideoStats().String(),
		workStatus.Status.String(),
		item.CompletedFormatted(),
		item.RemainingFormatted(),
		errString,
		table.UserData{Data: item},
	}
}
