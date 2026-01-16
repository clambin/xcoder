package tui

import (
	"codeberg.org/clambin/bubbles/table"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

type KeyMap struct {
	QueueViewer QueueViewerKeyMap
	LogViewer   LogViewerKeyMap
	Quit        key.Binding
	ShowLogs    key.Binding
}

func DefaultKeyMap() KeyMap {
	return KeyMap{
		QueueViewer: DefaultQueueViewerKeyMap(),
		LogViewer:   DefaultLogViewerKeyMap(),
		Quit:        key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
		ShowLogs:    key.NewBinding(key.WithKeys("l"), key.WithHelp("l", "show logs")),
	}
}

// QueueViewerKeyMap defines the keybindings for the queueViewer
type QueueViewerKeyMap struct {
	FilterTableKeyMap table.FilterTableKeyMap
	MediaFilterKeyMap MediaFilterKeyMap
	ActivateQueue     key.Binding
	Convert           key.Binding
	ShowFullPath      key.Binding
}

func DefaultQueueViewerKeyMap() QueueViewerKeyMap {
	return QueueViewerKeyMap{
		FilterTableKeyMap: table.DefaultFilterTableKeyMap(),
		MediaFilterKeyMap: DefaultMediaFilterKeyMap(),
		ActivateQueue:     key.NewBinding(key.WithKeys("p"), key.WithHelp("p", "toggle queue active state")),
		Convert:           key.NewBinding(key.WithKeys(tea.KeyEnter.String()), key.WithHelp(tea.KeyEnter.String(), "convert selected item")),
		ShowFullPath:      key.NewBinding(key.WithKeys("f"), key.WithHelp("f", "show full path")),
	}
}

var _ help.KeyMap = &MediaFilterKeyMap{}

// MediaFilterKeyMap defines the keybindings for the media filter
type MediaFilterKeyMap struct {
	ShowSkippedFiles   key.Binding
	ShowRejectedFiles  key.Binding
	ShowConvertedFiles key.Binding
}

func DefaultMediaFilterKeyMap() MediaFilterKeyMap {
	return MediaFilterKeyMap{
		ShowSkippedFiles:   key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "show/hide skipped files")),
		ShowRejectedFiles:  key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "show/hide rejected files")),
		ShowConvertedFiles: key.NewBinding(key.WithKeys("c"), key.WithHelp("c", "show/hide converted files")),
	}
}

func (m MediaFilterKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{m.ShowSkippedFiles, m.ShowRejectedFiles, m.ShowConvertedFiles}
}

func (m MediaFilterKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{m.ShortHelp()}
}

var _ help.KeyMap = LogViewerKeyMap{}

type LogViewerKeyMap struct {
	WordWrap   key.Binding
	AutoScroll key.Binding
	CloseLogs  key.Binding
}

func DefaultLogViewerKeyMap() LogViewerKeyMap {
	return LogViewerKeyMap{
		WordWrap:   key.NewBinding(key.WithKeys("w"), key.WithHelp("w", "wrap words")),
		AutoScroll: key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "auto scroll")),
		CloseLogs:  key.NewBinding(key.WithKeys(tea.KeyEscape.String()), key.WithHelp(tea.KeyEscape.String(), "close logs")),
	}
}

func (l LogViewerKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{l.WordWrap, l.AutoScroll, l.CloseLogs}
}

func (l LogViewerKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{l.ShortHelp()}
}
