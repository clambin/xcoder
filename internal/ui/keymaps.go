package ui

import (
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"codeberg.org/clambin/bubbles/table"
)

type KeyMap struct {
	RootKeyMap
	LogViewerKeyMap
	MediaViewerKeyMap
}

func DefaultKeyMap() KeyMap {
	return KeyMap{
		RootKeyMap: RootKeyMap{
			Quit: key.NewBinding(
				key.WithKeys("q"),
				key.WithHelp("q", "quit application"),
			),
			Logs: key.NewBinding(
				key.WithKeys("l"),
				key.WithHelp("l", "toggle logs"),
			),
			Help: key.NewBinding(
				key.WithKeys("?", "f1"),
				key.WithHelp("?/f1", "toggle help"),
			),
		},
		LogViewerKeyMap: LogViewerKeyMap{
			WordWrap:   key.NewBinding(key.WithKeys("w"), key.WithHelp("w", "wrap words")),
			AutoScroll: key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "auto scroll")),
			CloseLogs:  key.NewBinding(key.WithKeys("esc", "r"), key.WithHelp("esc", "close logs")),
		},
		MediaViewerKeyMap: MediaViewerKeyMap{
			ShowFullPath:       key.NewBinding(key.WithKeys("f"), key.WithHelp("f", "toggle full file path")),
			HideSkippedFiles:   key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "toggle skipped files")),
			HideRejectedFiles:  key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "toggle rejected files")),
			HideConvertedFiles: key.NewBinding(key.WithKeys("c"), key.WithHelp("c", "toggle completed files")),
			ConvertSelected:    key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "convert selected file")),
			AutoProcess:        key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "activate batch processing")),
			FilterTableKeyMap:  table.DefaultFilterTableKeyMap(),
		},
	}
}

var (
	_ help.KeyMap = RootKeyMap{}
	_ help.KeyMap = LogViewerKeyMap{}
	_ help.KeyMap = MediaViewerKeyMap{}
)

type RootKeyMap struct {
	Quit key.Binding
	Logs key.Binding
	Help key.Binding
}

func (r RootKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{r.Help}
}

func (r RootKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{r.Quit, r.Help, r.Logs}}
}

type LogViewerKeyMap struct {
	WordWrap   key.Binding
	AutoScroll key.Binding
	CloseLogs  key.Binding
}

func (l LogViewerKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{l.WordWrap, l.AutoScroll}
}

func (l LogViewerKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{l.WordWrap, l.AutoScroll, l.CloseLogs}}
}

type MediaViewerKeyMap struct {
	ShowFullPath       key.Binding
	HideSkippedFiles   key.Binding
	HideRejectedFiles  key.Binding
	HideConvertedFiles key.Binding
	ConvertSelected    key.Binding
	AutoProcess        key.Binding
	table.FilterTableKeyMap
}

func (l MediaViewerKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		l.HideSkippedFiles,
		l.HideRejectedFiles,
		l.HideConvertedFiles,
	}
}

func (l MediaViewerKeyMap) FullHelp() [][]key.Binding {
	return append([][]key.Binding{{
		l.ShowFullPath,
		l.HideSkippedFiles,
		l.HideRejectedFiles,
		l.HideConvertedFiles,
		l.ConvertSelected,
		l.AutoProcess,
	}}, l.FilterKeyMap.FullHelp()...)
}
