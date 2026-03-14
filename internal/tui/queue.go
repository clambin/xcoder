package tui

import (
	"path/filepath"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"codeberg.org/clambin/bubbles/frame"
	"codeberg.org/clambin/bubbles/table"
	"github.com/clambin/xcoder/internal/pipeline"
)

const title = "media files"

var (
	columns = []table.Column{
		{Name: "SOURCE"},
		{Name: "SOURCE STATS", Width: 25},
		{Name: "TARGET STATS", Width: 25},
		{Name: "STATUS", Width: 10, CellStyle: table.CellStyle{Style: lipgloss.NewStyle().Transform(table.StringStyler(statusColors))}},
		{Name: "COMPLETED", Width: 10, CellStyle: table.CellStyle{Style: lipgloss.NewStyle().Align(lipgloss.Right)}},
		{Name: "REMAINING", Width: 10, CellStyle: table.CellStyle{Style: lipgloss.NewStyle().Align(lipgloss.Right)}},
		{Name: "ERROR", Width: 40},
	}
)

var _ Queue = (*pipeline.Queue)(nil)

// Queue is the interface for a pipeline.Queue.
type Queue interface {
	Queue(item *pipeline.WorkItem)
	SetActive(active bool)
	All() []*pipeline.WorkItem
	Active() bool
	Stats() map[pipeline.Status]int
}

type queueViewer struct {
	frameStyle       frame.Style
	mediaFilterStyle lipgloss.Style
	queue            Queue
	keyMap           QueueViewerKeyMap
	mediaFilter      mediaFilter
	table.FilterTable
	textFilterOn bool
	showFullPath bool
}

func newQueueViewer(queue Queue, styles QueueViewerStyles, keyMap QueueViewerKeyMap) queueViewer {
	return queueViewer{
		FilterTable:      table.NewFilterTable().Columns(columns).Styles(styles.Table).KeyMap(keyMap.FilterTableKeyMap),
		queue:            queue,
		keyMap:           keyMap,
		mediaFilter:      mediaFilter{KeyMap: keyMap.MediaFilterKeyMap},
		mediaFilterStyle: styles.MediaFilter,
		frameStyle:       styles.Frame,
	}
}

func (q queueViewer) Update(msg tea.Msg) (queueViewer, tea.Cmd) {
	// fmt.Printf("msg: %#+v\n", msg)

	var cmd tea.Cmd
	switch msg := msg.(type) {
	case table.FilterStateChangeMsg:
		// text filter changed.  if on, we need to route all keys to the table.
		q.textFilterOn = msg.State
	case refreshUIMsg:
		// refresh the table. this is done in a cmd, so it doesn't block the UI loop.
		cmd = loadTableCmd(q.queue.All(), q.mediaFilter.mediaFilterState, q.showFullPath)
	case setRowsMsg:
		q.FilterTable = q.Rows(msg.rows)
	case tea.KeyMsg:
		// if the text filter is active, it receives all inputs.
		if q.textFilterOn {
			q.FilterTable, cmd = q.FilterTable.Update(msg)
			break
		}
		switch {
		case key.Matches(msg, q.keyMap.ActivateQueue):
			// toggle queue active state
			q.queue.SetActive(!q.queue.Active())
		case key.Matches(msg, q.keyMap.Convert):
			if row := q.SelectedRow(); row != nil {
				q.queue.Queue(row[len(row)-1].(table.UserData).Data.(*pipeline.WorkItem))
				cmd = func() tea.Msg { return refreshUIMsg{} }
			}
		case key.Matches(msg, q.keyMap.ShowFullPath):
			q.showFullPath = !q.showFullPath
			// refresh the table
			cmd = func() tea.Msg { return refreshUIMsg{} }
		default:
			// route key to mediaFilter
			if q.mediaFilter, cmd = q.mediaFilter.Update(msg); cmd == nil {
				// if no action, route to table
				q.FilterTable, cmd = q.FilterTable.Update(msg)
			}
		}
	default:
		// any other message is passed to the table
		q.FilterTable, cmd = q.FilterTable.Update(msg)
	}
	return q, cmd
}

func (q queueViewer) View() string {
	frameTitle := title
	if f := q.mediaFilter.View(); f != "" {
		frameTitle += " (" + q.mediaFilterStyle.Render(f) + ")"
	}
	return frame.Render(frameTitle, lipgloss.Center, q.frameStyle, q.FilterTable.View())
}

func (q queueViewer) SetSize(width, height int) queueViewer {
	borderWidth := q.frameStyle.Border.GetHorizontalBorderSize()
	borderHeight := q.frameStyle.Border.GetVerticalBorderSize()
	q.FilterTable = q.Size(width-borderWidth, height-borderHeight)
	return q
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
