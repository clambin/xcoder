package ui

import "github.com/rivo/tview"

var logViewerShortCuts = shortcutsPage{
	{shortcut{key: "l", description: "close logs"}},
}

type LogViewer struct {
	*tview.TextView
}

func newLogViewer() *LogViewer {
	v := LogViewer{TextView: tview.NewTextView()}
	v.
		SetBorder(true).
		SetTitle(" logs ").
		SetTitleAlign(tview.AlignCenter).
		SetBorderPadding(0, 0, 1, 1)
	v.
		SetScrollable(true).
		ScrollToEnd().
		SetWrap(true)
	return &v
}
