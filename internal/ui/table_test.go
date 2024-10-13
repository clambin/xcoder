package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
)

func TestTable_Update(t *testing.T) {
	dataSource := fakeDataSource{rows: []string{"0", "1", "2", "3", "4"}}
	table := NewTable(&dataSource)

	for len(dataSource.rows) > 0 {
		table.Update()
		assert.Equal(t, len(dataSource.rows)+1, table.GetRowCount())
		for i, r := range dataSource.rows {
			assert.Equal(t, r, table.Table.GetCell(i+1, 0).Text)
		}
		dataSource.rows = dataSource.rows[1:]
	}
}

// Current:
// BenchmarkTable_Update-16            9244            109370 ns/op          272835 B/op       2003 allocs/op
// With cellPool:
// BenchmarkTable_Update-16           15936             75340 ns/op           33383 B/op       1003 allocs/op
// With Header optimization:
// BenchmarkTable_Update-16           16092             74616 ns/op           33157 B/op       1001 allocs/op
func BenchmarkTable_Update(b *testing.B) {
	dataSource := fakeDataSource{rows: make([]string, 1000)}
	for i := range len(dataSource.rows) {
		dataSource.rows[i] = strconv.Itoa(i)
	}
	table := NewTable(&dataSource)
	b.ResetTimer()
	for range b.N {
		table.Update()
	}
}

var _ DataSource = &fakeDataSource{}

type fakeDataSource struct {
	rows []string
}

func (f fakeDataSource) Header() []*tview.TableCell {
	return []*tview.TableCell{tview.NewTableCell("Status")}
}

func (f fakeDataSource) Update() Update {
	rows := make([][]*tview.TableCell, len(f.rows))
	for r := range f.rows {
		rows[r] = []*tview.TableCell{getTableCell(f.rows[r], tcell.ColorWhite, tcell.ColorBlack, tview.AlignLeft)}
	}
	return Update{
		Rows:   rows,
		Title:  "",
		Reload: false,
	}
}

func (f fakeDataSource) HandleInput(_ *tcell.EventKey) *tcell.EventKey {
	panic("implement me")
}

//////////////////////////////////////////////////////////////////////////////////////////////////////

func TestCellAllocator(t *testing.T) {
	c := getTableCell("foo", tcell.ColorWhite, tcell.ColorBlack, tview.AlignRight)
	assert.Equal(t, "foo", c.Text)
	fg, bg, _ := c.Style.Decompose()
	assert.Equal(t, tcell.ColorWhite, fg)
	assert.Equal(t, tcell.ColorBlack, bg)
	assert.Equal(t, tview.AlignRight, c.Align)
	putTableCell(c)
	c = getTableCell("bar", tcell.ColorRed, tcell.ColorBlue, tview.AlignRight)
	fg, bg, _ = c.Style.Decompose()
	assert.Equal(t, "bar", c.Text)
	assert.Equal(t, tcell.ColorRed, fg)
	assert.Equal(t, tcell.ColorBlue, bg)
	assert.Equal(t, tview.AlignRight, c.Align)

}

func BenchmarkCellPool(b *testing.B) {
	const cellCount = 100
	b.Run("pool", func(b *testing.B) {
		for range b.N {
			cells := make([]*tview.TableCell, cellCount)
			for i := range cells {
				label := strconv.Itoa(i)
				cells[i] = getTableCell(label, tcell.ColorWhite, tcell.ColorBlack, tview.AlignLeft)
				if cells[i] == nil {
					b.Fatal("no cell allocated")
				}
				if cells[i].Text != label {
					b.Fatal("cell allocated does not match " + label)
				}
			}
			for _, cell := range cells {
				putTableCell(cell)
			}
		}
	})
	b.Run("direct", func(b *testing.B) {
		for range b.N {
			cells := make([]*tview.TableCell, cellCount)
			for i := range cells {
				label := strconv.Itoa(i)
				cells[i] = tview.NewTableCell(label).SetTextColor(tcell.ColorWhite).SetBackgroundColor(tcell.ColorBlack).SetAlign(tview.AlignLeft)
				if cells[i] == nil {
					b.Fatal("no cell allocated")
				}
				if cells[i].Text != label {
					b.Fatal("cell allocated does not match " + label)
				}
			}
		}
	})
}
