package ui

import (
	"github.com/clambin/go-common/set"
	"github.com/clambin/videoConvertor/internal/worklist"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
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
	v.SetInputCapture(v.handleInput)
	v.Table.
		SetEvaluateAllRows(true).
		SetFixed(1, 0).
		SetSelectable(true, false).
		SetBorder(true).
		SetBorderPadding(0, 0, 1, 1)
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
	for _, entry := range list {
		status, err := entry.Status()
		if v.filters.on(status) {
			continue
		}
		rowCount++
		source := entry.Source
		if !v.fullName.Load() {
			source = filepath.Base(source)
		}
		v.SetCell(rowCount, 0, tview.NewTableCell(source).SetReference(entry))
		v.SetCell(rowCount, 1, tview.NewTableCell(entry.SourceVideoStats().String()))
		v.SetCell(rowCount, 2, tview.NewTableCell(entry.TargetVideoStats().String()))
		statusColor, ok := tableColorStatus[status]
		if !ok {
			statusColor = tview.Styles.PrimaryTextColor
		}
		v.SetCell(rowCount, 3, tview.NewTableCell(status.String()).SetTextColor(statusColor))
		var progress string
		var remaining string
		if status == worklist.Converting {
			if p := entry.Progress.Completed(); p > 0 {
				progress = strconv.FormatFloat(100*p, 'f', 1, 64) + "%"
			}
			if r := entry.Remaining(); r >= 0 {
				remaining = formatDuration(r)
			}
		}
		v.SetCell(rowCount, 4, tview.NewTableCell(progress).SetAlign(tview.AlignRight))
		v.SetCell(rowCount, 5, tview.NewTableCell(remaining).SetAlign(tview.AlignRight))
		var errString string
		if err != nil {
			errString = err.Error()
		}
		v.SetCell(rowCount, 6, tview.NewTableCell(errString))
	}
	v.Table.SetTitle(v.title(list, rowCount))

	if currentRow, _ := v.Table.GetSelection(); currentRow == 0 && rowCount > 0 {
		v.Table.ScrollToBeginning()
	}
}

func (v *workListViewer) title(list []*worklist.WorkItem, rows int) string {
	title := "files"
	var filtered bool
	if f := v.filters.list(); len(f) > 0 {
		fs := make([]string, len(f))
		for i, e := range f {
			fs[i] = e.String()
		}
		title += " (filtered: " + strings.Join(fs, ", ") + ")"
		filtered = len(fs) > 0
	}
	if filtered {
		title += " [" + strconv.Itoa(rows) + "/" + strconv.Itoa(len(list)) + "]"

	} else {
		title += " [" + strconv.Itoa(rows) + "]"
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

/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func formatDuration(d time.Duration) string {
	var output string
	var days int
	for d >= 24*time.Hour {
		days++
		d -= 24 * time.Hour
	}
	if days > 0 {
		output = strconv.Itoa(days) + "d"
	}
	if hours := int(d.Hours()); hours > 0 {
		output += strconv.Itoa(hours) + "h"
		d -= time.Duration(hours) * time.Hour
	}
	if minutes := int(d.Minutes()); minutes > 0 {
		output += strconv.Itoa(minutes) + "m"
		d -= time.Duration(minutes) * time.Minute
	}
	if seconds := int(d.Seconds()); seconds > 0 {
		output += strconv.Itoa(seconds) + "s"
	}
	return output
}
