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
func BenchmarkTable_Update(b *testing.B) {
	dataSource := fakeDataSource{rows: make([]string, 1000)}
	for i := range len(dataSource.rows) {
		dataSource.rows[i] = strconv.Itoa(i)
	}
	table := NewTable(&dataSource)
	b.ResetTimer()
	for range b.N {
		table.Update()
		//dataSource.rows = dataSource.rows[1:]
	}
}

var _ DataSource = &fakeDataSource{}

type fakeDataSource struct {
	rows []string
}

func (f fakeDataSource) Update() Update {
	hdr := []*tview.TableCell{tview.NewTableCell("Status")}
	rows := make([][]*tview.TableCell, len(f.rows))
	for r := range f.rows {
		rows[r] = []*tview.TableCell{getTableCell(f.rows[r], tcell.ColorWhite, tcell.ColorBlack, tview.AlignLeft)}
	}
	return Update{
		Headers: hdr,
		Rows:    rows,
		Title:   "",
		Reload:  false,
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

func BenchmarkCellAllocator(b *testing.B) {
	b.Run("pool", func(b *testing.B) {
		for range b.N {
			cell := getTableCell("foo", tcell.ColorWhite, tcell.ColorBlack, tview.AlignLeft)
			if cell == nil {
				b.Fatal("no cell allocated")
			}
			if cell.Text != "foo" {
				b.Fatal("cell allocated does not match foo")
			}
			putTableCell(cell)
		}
	})
	b.Run("direct", func(b *testing.B) {
		for range b.N {
			if cell := tview.NewTableCell("foo").SetTextColor(tcell.ColorWhite).SetBackgroundColor(tcell.ColorBlack).SetAlign(tview.AlignLeft); cell == nil {
				b.Fatal("no cell allocated")
			}
		}
	})
}
