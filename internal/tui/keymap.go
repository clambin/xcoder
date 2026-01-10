package tui

import (
	"codeberg.org/clambin/bubbles/table"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

type KeyMap struct {
	Controller  ControllerKeyMap
	Filter      FilterKeyMap
	LogViewer   logViewerKeyMap
	QueueViewer queueViewerKeyMap
}

func defaultKeyMap() KeyMap {
	return KeyMap{
		Controller:  defaultControllerKeyMap(),
		Filter:      defaultFilterKeyMap(),
		LogViewer:   defaultLogViewerKeyMap(),
		QueueViewer: defaultQueueViewerKeyMap(),
	}
}

var _ help.KeyMap = ControllerKeyMap{}

type ControllerKeyMap struct {
	Quit     key.Binding
	Activate key.Binding
	Convert  key.Binding
	FullPath key.Binding
	ShowLogs key.Binding
}

func defaultControllerKeyMap() ControllerKeyMap {
	return ControllerKeyMap{
		Quit:     key.NewBinding(key.WithKeys(tea.KeyCtrlC.String(), "q"), key.WithHelp("q/ctrl+c", "quit")),
		Activate: key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "activate batch transcoding")),
		Convert:  key.NewBinding(key.WithKeys(tea.KeyEnter.String()), key.WithHelp("enter", "transcode selected item")),
		FullPath: key.NewBinding(key.WithKeys("f"), key.WithHelp("f", "show full path")),
		ShowLogs: key.NewBinding(key.WithKeys("l"), key.WithHelp("l", "show logs")),
	}
}

func (k ControllerKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.Quit,
		k.Activate,
		k.Convert,
	}
}

func (k ControllerKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		k.ShortHelp(),            // General
		{k.FullPath, k.ShowLogs}, // View
	}
}

var _ help.KeyMap = FilterKeyMap{}

type FilterKeyMap struct {
	ShowSkippedFiles   key.Binding
	ShowRejectedFiles  key.Binding
	ShowConvertedFiles key.Binding
}

func defaultFilterKeyMap() FilterKeyMap {
	return FilterKeyMap{
		ShowSkippedFiles:   key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "show/hide skipped files")),
		ShowRejectedFiles:  key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "show/hide rejected files")),
		ShowConvertedFiles: key.NewBinding(key.WithKeys("c"), key.WithHelp("c", "show/hide converted files"))}
}

func (f FilterKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		f.ShowSkippedFiles,
		f.ShowRejectedFiles,
		f.ShowConvertedFiles,
	}
}

func (f FilterKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{f.ShortHelp()}
}

var _ help.KeyMap = logViewerKeyMap{}

// logViewerKeyMap contains the key bindings for the logViewer.
type logViewerKeyMap struct {
	WordWrap   key.Binding
	AutoScroll key.Binding
	CloseLogs  key.Binding
}

func defaultLogViewerKeyMap() logViewerKeyMap {
	return logViewerKeyMap{
		WordWrap:   key.NewBinding(key.WithKeys("w"), key.WithHelp("w", "wrap words")),
		AutoScroll: key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "auto scroll")),
		CloseLogs:  key.NewBinding(key.WithKeys(tea.KeyEscape.String(), "l"), key.WithHelp(tea.KeyEscape.String()+"/l", "close logs")),
	}
}

func (l logViewerKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{l.CloseLogs, l.AutoScroll, l.CloseLogs}
}

func (l logViewerKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{l.ShortHelp()}
}

// queueViewerKeyMap contains the key bindings for the queueViewer.
type queueViewerKeyMap struct {
	FilterTableKeyMap table.FilterTableKeyMap
	FullPath          key.Binding
}

func defaultQueueViewerKeyMap() queueViewerKeyMap {
	return queueViewerKeyMap{
		FilterTableKeyMap: table.DefaultFilterTableKeyMap(),
		FullPath:          key.NewBinding(key.WithKeys("f"), key.WithHelp("f", "show full path")),
	}
}
