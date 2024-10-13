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
// BenchmarkTable_Update-16                    9350            108381 ns/op          272835 B/op       2003 allocs/op
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
		rows[r] = []*tview.TableCell{tview.NewTableCell(f.rows[r])}
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
