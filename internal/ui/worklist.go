package ui

import (
	"github.com/clambin/go-common/set"
	"github.com/clambin/videoConvertor/internal/worklist"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
)

var workListShortCuts = shortcutsPage{
	{
		shortcut{key: "p", description: "enable/disable processing"},
		shortcut{key: "enter", description: "convert selected file"},
		shortcut{key: "l", description: "show logs"},
	},
	{
		shortcut{key: "s", description: "show/hide skipped files"},
		shortcut{key: "r", description: "show/hide rejected files"},
		shortcut{key: "c", description: "show/hide converted files"},
		shortcut{key: "f", description: "show/hide full path name"},
	},
}

type workListViewer struct {
	*Table
	list *worklist.WorkList
	DataSource
}

func newWorkListViewer(list *worklist.WorkList) *workListViewer {
	dataSource := &workItems{
		list:    list,
		filters: filters{statuses: set.New[worklist.WorkStatus]()},
	}
	return &workListViewer{
		Table:      NewTable(dataSource),
		list:       list,
		DataSource: dataSource,
	}
}

func (v *workListViewer) refresh() {
	v.Table.Update()
}

func (v *workListViewer) handleInput(event *tcell.EventKey) *tcell.EventKey {
	if v.DataSource.HandleInput(event) == nil {
		return nil
	}
	switch event.Key() {
	case tcell.KeyRune:
		switch event.Rune() {
		case 'p':
			v.list.ToggleActive()
			return nil
		default:
			return event
		}
	case tcell.KeyEnter:
		row, _ := v.Table.GetSelection()
		item := v.Table.GetCell(row, 0).GetReference().(*worklist.WorkItem)
		if status, _ := item.Status(); status == worklist.Inspected || status == worklist.Failed {
			v.list.Queue(item)
		}
		return nil
	default:
		return event
	}
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type filters struct {
	statuses set.Set[worklist.WorkStatus]
	changed  bool
	lock     sync.RWMutex
}

func (f *filters) toggle(statuses ...worklist.WorkStatus) {
	f.lock.Lock()
	defer f.lock.Unlock()
	for _, status := range statuses {
		if f.statuses.Contains(status) {
			f.statuses.Remove(status)
		} else {
			f.statuses.Add(status)
		}
		f.changed = true
	}
}

func (f *filters) on(status worklist.WorkStatus) bool {
	f.lock.RLock()
	defer f.lock.RUnlock()
	return f.statuses.Contains(status)
}

func (f *filters) list() []worklist.WorkStatus {
	f.lock.RLock()
	defer f.lock.RUnlock()
	return f.statuses.ListOrdered()
}

func (f *filters) Format() string {
	filtered := f.list()
	if len(filtered) == 0 {
		return ""
	}
	fs := make([]string, len(filtered))
	for i, e := range filtered {
		fs[i] = e.String()
	}
	slices.Sort(fs)
	return strings.Join(fs, ", ")
}

func (f *filters) updated() bool {
	f.lock.Lock()
	defer f.lock.Unlock()
	changed := f.changed
	f.changed = false
	return changed
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func colorStatus(item *worklist.WorkItem) tcell.Color {
	status, _ := item.Status()
	if statusColor, ok := tableColorStatus[status]; ok {
		return statusColor
	}
	return tview.Styles.PrimaryTextColor
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////

var _ DataSource = &workItems{}

type workItems struct {
	list *worklist.WorkList
	filters
	fullName atomic.Bool
}

var headerCells = []*tview.TableCell{
	tview.NewTableCell(padString("SOURCE", 100)).SetSelectable(false).SetTextColor(tview.Styles.SecondaryTextColor).SetAlign(tview.AlignLeft),
	tview.NewTableCell("SOURCE STATS").SetSelectable(false).SetTextColor(tview.Styles.SecondaryTextColor).SetAlign(tview.AlignLeft),
	tview.NewTableCell("TARGET STATS").SetSelectable(false).SetTextColor(tview.Styles.SecondaryTextColor).SetAlign(tview.AlignLeft),
	tview.NewTableCell(padString("STATUS", 9)).SetSelectable(false).SetTextColor(tview.Styles.SecondaryTextColor).SetAlign(tview.AlignLeft),
	tview.NewTableCell("COMPLETED").SetSelectable(false).SetTextColor(tview.Styles.SecondaryTextColor).SetAlign(tview.AlignRight),
	tview.NewTableCell("REMAINING").SetSelectable(false).SetTextColor(tview.Styles.SecondaryTextColor).SetAlign(tview.AlignRight),
	tview.NewTableCell(padString("ERROR", 20)).SetSelectable(false).SetTextColor(tview.Styles.SecondaryTextColor).SetAlign(tview.AlignLeft),
}

func padString(s string, width int) string {
	if toPad := width - len(s); toPad > 0 {
		s += strings.Repeat(" ", toPad)
	}
	return s
}

func (w *workItems) Update() Update {
	list := w.list.List()
	update := Update{
		Headers: headerCells,
		Rows:    make([][]*tview.TableCell, 0, len(list)),
		Reload:  w.filters.updated(),
	}
	for _, item := range list {
		if row := w.buildRow(item); row != nil {
			update.Rows = append(update.Rows, row)
		}
	}
	update.Title = w.title(len(list), len(update.Rows))

	return update
}

func (w *workItems) buildRow(item *worklist.WorkItem) []*tview.TableCell {
	status, err := item.Status()
	if w.filters.on(status) {
		return nil
	}
	source := item.Source
	if !w.fullName.Load() {
		source = filepath.Base(source)
	}
	var errString string
	if err != nil {
		errString = err.Error()
	}
	return []*tview.TableCell{
		tview.NewTableCell(source).SetReference(item),
		tview.NewTableCell(item.SourceVideoStats().String()),
		tview.NewTableCell(item.TargetVideoStats().String()),
		tview.NewTableCell(status.String()).SetTextColor(colorStatus(item)),
		tview.NewTableCell(item.CompletedFormatted()).SetAlign(tview.AlignRight),
		tview.NewTableCell(item.RemainingFormatted()).SetAlign(tview.AlignRight),
		tview.NewTableCell(errString),
	}
}

func (w *workItems) title(itemCount, rowCount int) string {
	title := "files"
	filtered := w.filters.Format()
	if filtered != "" {
		title += " (filtered: " + filtered + ")[" + strconv.Itoa(rowCount) + "/" + strconv.Itoa(itemCount) + "]"
	} else {
		title += " [" + strconv.Itoa(rowCount) + "]"
	}
	return " " + title + " "
}

func (w *workItems) HandleInput(event *tcell.EventKey) *tcell.EventKey {
	if event.Key() != tcell.KeyRune {
		return event
	}
	switch event.Rune() {
	case 's':
		w.filters.toggle(worklist.Skipped)
		return nil
	case 'c':
		w.filters.toggle(worklist.Converted)
		return nil
	case 'r':
		w.filters.toggle(worklist.Rejected)
		return nil
	case 'f':
		w.fullName.Store(!w.fullName.Load())
		return nil
	}
	return event
}
