package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"sync"
)

type DataSource interface {
	Update() Update
	HandleInput(event *tcell.EventKey) *tcell.EventKey
}

type Update struct {
	Headers []*tview.TableCell
	Rows    [][]*tview.TableCell
	Title   string
	Reload  bool
}

type Table struct {
	*tview.Table
	DataSource
}

func NewTable(source DataSource) *Table {
	t := Table{
		Table:      tview.NewTable(),
		DataSource: source,
	}
	t.Table.SetEvaluateAllRows(true).
		SetFixed(1, 0).
		SetSelectable(true, false).
		Select(1, 0).
		SetBorder(true).
		SetBorderPadding(0, 0, 1, 1).
		SetInputCapture(t.handleInput)
	return &t
}

func (t *Table) Update() {
	update := t.DataSource.Update()
	t.Table.SetTitle(update.Title)
	for i, h := range update.Headers {
		t.Table.SetCell(0, i, h)
	}
	for i, row := range update.Rows {
		for j, cell := range row {
			putTableCell(t.Table.GetCell(i+1, j))
			t.Table.SetCell(i+1, j, cell)
		}
	}
	t.trimRows(len(update.Rows) + 1)
	if update.Reload {
		t.Table.Select(1, 0)
		t.Table.ScrollToBeginning()
	}
}

func (t *Table) trimRows(rows int) {
	for t.Table.GetRowCount() > rows {
		r := t.Table.GetRowCount() - 1
		for c := range t.Table.GetColumnCount() {
			putTableCell(t.Table.GetCell(r, c))
		}
		t.Table.RemoveRow(r)
	}
}

func (t *Table) handleInput(event *tcell.EventKey) *tcell.EventKey {
	return t.DataSource.HandleInput(event)
}

//////////////////////////////////////////////////////////////////////////////////////////////////////

// cellPool reduces memory allocations for tview.TableCell objects
var cellPool = sync.Pool{
	New: func() any {
		return tview.NewTableCell("")
	},
}

func getTableCell(label string, fgColor, bgColor tcell.Color, align int) *tview.TableCell {
	cell := cellPool.Get().(*tview.TableCell)
	cell.SetText(label).SetTextColor(fgColor).SetBackgroundColor(bgColor).SetAlign(align)
	return cell
}

func putTableCell(cell *tview.TableCell) {
	cellPool.Put(cell)
}
