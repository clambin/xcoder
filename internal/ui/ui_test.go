package ui

import (
	"github.com/clambin/videoConvertor/internal/configuration"
	"github.com/clambin/videoConvertor/internal/worklist"
	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestUI(t *testing.T) {
	var list worklist.WorkList
	list.Add("foo")
	list.List()[0].Done(worklist.Skipped, nil)

	cfg := configuration.Configuration{Input: "/foo", ProfileName: "foo"}
	ui := New(&list, cfg)
	ui.refresh()

	// validate header - configuration
	const wantConfigText = ` Base directory:   /foo
 Profile:          foo
 Remove source:    off
 Overwrite target: off
`
	assert.Equal(t, wantConfigText, ui.header.configPane.GetText(true))
	// validate header - status
	const wantStatusText = ` Converting    : off
`
	assert.Equal(t, wantStatusText, ui.header.statusPane.GetText(true))

	// default view: table is unfiltered
	assert.Equal(t, 2, ui.workListViewer.Table.GetRowCount())
	assert.Equal(t, "foo", ui.workListViewer.Table.GetCell(1, 0).Text)

	// validate that filters work
	ui.handleInput(tcell.NewEventKey(tcell.KeyRune, 's', tcell.ModNone))
	ui.refresh()
	assert.Equal(t, 1, ui.workListViewer.Table.GetRowCount())
	ui.handleInput(tcell.NewEventKey(tcell.KeyRune, 's', tcell.ModNone))
	ui.refresh()
	assert.Equal(t, 2, ui.workListViewer.Table.GetRowCount())

	// verify switching between different pages
	front, _ := ui.GetFrontPage()
	assert.Equal(t, "worklist", front)
	ui.handleInput(tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone))
	ui.refresh()
	front, _ = ui.GetFrontPage()
	assert.Equal(t, "logs", front)
	ui.handleInput(tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone))
	ui.refresh()
	front, _ = ui.GetFrontPage()
	assert.Equal(t, "worklist", front)

	// Request a file to be converted
	ui.Select(1, 0)
	list.List()[0].Done(worklist.Inspected, nil)
	ui.handleInput(tcell.NewEventKey(tcell.KeyEnter, rune(tcell.KeyEnter), tcell.ModNone))
	item := list.NextToConvert()
	require.NotNil(t, item)
	assert.Equal(t, "foo", item.Source)
}
