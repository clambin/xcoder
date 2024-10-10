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
		shortcut{key: "l", description: "show logs"},
		shortcut{key: "enter", description: "convert selected file"},
	},
	{
		shortcut{key: "s", description: "show/hide skipped files"},
		shortcut{key: "r", description: "show/hide rejected files"},
		shortcut{key: "c", description: "show/hide converted files"},
		shortcut{key: "f", description: "show/hide full path name"},
	},
}

type workListViewer struct {
	*tview.Table
	list     *worklist.WorkList
	filters  *filters
	fullName atomic.Bool
}

func newWorkListViewer(list *worklist.WorkList) *workListViewer {
	v := workListViewer{
		list:    list,
		Table:   tview.NewTable(),
		filters: &filters{statuses: set.New[worklist.WorkStatus]()},
	}
	v.
		SetEvaluateAllRows(true).
		SetFixed(1, 0).
		SetSelectable(true, false).
		Select(1, 0).
		SetBorder(true).
		SetBorderPadding(0, 0, 1, 1).
		SetInputCapture(v.handleInput)
	return &v
}

type column struct {
	name        string
	orientation int
}

var columns = []column{
	{"SOURCE", tview.AlignLeft},
	{"SOURCE STATS", tview.AlignLeft},
	{"TARGET STATS", tview.AlignLeft},
	{"STATUS    ", tview.AlignLeft},
	{"COMPLETED", tview.AlignRight},
	{"REMAINING", tview.AlignRight},
	{"ERROR", tview.AlignLeft},
}

func (v *workListViewer) refresh() {
	selectedItem := v.selectedItem()
	v.Clear()
	list := v.list.List()

	for i, col := range columns {
		v.SetCell(0, i, tview.NewTableCell(col.name).
			SetAlign(col.orientation).
			SetSelectable(false).
			SetTextColor(tview.Styles.SecondaryTextColor),
		)
	}
	var rowCount int
	for _, item := range list {
		status, err := item.Status()
		if v.filters.on(status) {
			continue
		}
		rowCount++
		source := item.Source
		if !v.fullName.Load() {
			source = filepath.Base(source)
		}
		v.SetCell(rowCount, 0, tview.NewTableCell(source).SetExpansion(10).SetReference(item))
		v.SetCell(rowCount, 1, tview.NewTableCell(item.SourceVideoStats().String()))
		v.SetCell(rowCount, 2, tview.NewTableCell(item.TargetVideoStats().String()))
		v.SetCell(rowCount, 3, tview.NewTableCell(status.String()).SetTextColor(colorStatus(item)))
		v.SetCell(rowCount, 4, tview.NewTableCell(item.CompletedFormatted()).SetAlign(tview.AlignRight))
		v.SetCell(rowCount, 5, tview.NewTableCell(item.RemainingFormatted()).SetAlign(tview.AlignRight))
		var errString string
		if err != nil {
			errString = err.Error()
		}
		v.SetCell(rowCount, 6, tview.NewTableCell(errString).SetExpansion(1))
	}
	v.Table.SetTitle(v.title(len(list), rowCount))

	v.selectRow(selectedItem)
}

func (v *workListViewer) selectedItem() *worklist.WorkItem {
	selectedRow, _ := v.Table.GetSelection()
	if item, ok := v.Table.GetCell(selectedRow, 0).GetReference().(*worklist.WorkItem); ok {
		return item
	}
	return nil
}

func (v *workListViewer) selectRow(item *worklist.WorkItem) {
	for r := range v.Table.GetRowCount() {
		if v.Table.GetCell(r, 0).GetReference() == item {
			v.Table.Select(r, 0)
			return
		}
	}
	v.ScrollToBeginning()
	if v.GetRowCount() > 1 {
		v.Select(1, 0)
	}
}

func (v *workListViewer) title(itemCount, rowCount int) string {
	title := "files"
	filtered := v.filters.Format()
	if filtered != "" {
		title += " (filtered: " + filtered + ")[" + strconv.Itoa(rowCount) + "/" + strconv.Itoa(itemCount) + "]"
	} else {
		title += " [" + strconv.Itoa(rowCount) + "]"
	}
	return " " + title + " "
}

func (v *workListViewer) handleInput(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyRune:
		switch event.Rune() {
		case 's':
			v.filters.toggle(worklist.Skipped)
			return nil
		case 'c':
			v.filters.toggle(worklist.Converted)
			return nil
		case 'r':
			v.filters.toggle(worklist.Rejected)
			return nil
		case 'f':
			v.fullName.Store(!v.fullName.Load())
			return nil
		case 'p':
			v.list.ToggleActive()
			return nil
		default:
			return event
		}
	case tcell.KeyEnter:
		row, _ := v.GetSelection()
		item := v.GetCell(row, 0).GetReference().(*worklist.WorkItem)
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

/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func colorStatus(item *worklist.WorkItem) tcell.Color {
	status, _ := item.Status()
	if statusColor, ok := tableColorStatus[status]; ok {
		return statusColor
	}
	return tview.Styles.PrimaryTextColor
}
