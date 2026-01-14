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
	table            tea.Model
	queue            Queue
	keyMap           QueueViewerKeyMap
	selectedRow      table.Row
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
	var cmd tea.Cmd
	// TODO: table doesn't lend itself well to SetSize() approach. Should have its own SetSize method
	// TODO: queueViewer should draw its frame (+title) itself. table shouldn't be concerned with frames
	q.table, cmd = q.table.Update(table.SetSizeMsg{Width: width, Height: height})
	for cmd != nil {
		q.table, cmd = q.table.Update(cmd())
	}
}

func (q *queueViewer) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case table.FilterStateChangeMsg:
		// text filter changed.  if on, we need to route all keys to the table.
		q.textFilterOn = msg.State
	case refreshUIMsg:
		// refresh the table. this is done in a cmd, so it doesn't block the UI loop.
		cmd = loadTableCmd(q.queue.All(), q.mediaFilter.mediaFilterState, q.showFullPath)
	case table.RowChangedMsg:
		// mark the selected row so we know which queue item to convert when the user hits <enter>.
		q.selectedRow = msg.Row
	case mediaFilterChangedMsg:
		// filter changed. change the table title
		newTitle := title
		if filter := q.mediaFilter.mediaFilterState.String(); filter != "" {
			newTitle += " (" + q.mediaFilterStyle.Render(filter) + ")"
		}
		cmd = func() tea.Msg { return table.SetTitleMsg{Title: newTitle} }
	case tea.KeyMsg:
		// if the text filter is active, it receives all inputs.
		if q.textFilterOn {
			q.table, cmd = q.table.Update(msg)
			break
		}
		switch {
		case key.Matches(msg, q.keyMap.ActivateQueue):
			// toggle queue active state
			q.queue.SetActive(!q.queue.Active())
		case key.Matches(msg, q.keyMap.Convert):
			if row := q.selectedRow; row != nil {
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
				q.table, cmd = q.table.Update(msg)
			}
		}
	default:
		// any other message is passed to the table
		q.table, cmd = q.table.Update(msg)
	}
	return cmd
}

func (q *queueViewer) View() string {
	// TODO: frame should be drawn here, not in table, so we can set the title directly.
	v := q.table.View()
	return v // q.table.View()
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
		return table.SetRowsMsg{Rows: rows}
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
