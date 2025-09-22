package tui

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

var (
	defaultKeyMap = ControllerKeyMap{
		Quit: key.NewBinding(
			key.WithKeys(tea.KeyCtrlC.String(), "q"),
			key.WithHelp("q/ctrl+c", "quit"),
		),
		Activate: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "activate batch transcoding"),
		),
		Convert: key.NewBinding(
			key.WithKeys(tea.KeyEnter.String()),
			key.WithHelp("enter", "transcode selected item"),
		),
		FullPath: key.NewBinding(
			key.WithKeys("f"),
			key.WithHelp("f", "show full path"),
		),
	}

	defaultFilterKeyMap = FilterKeyMap{
		ShowSkippedFiles: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "show/hide skipped files"),
		),
		ShowRejectedFiles: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "show/hide rejected files"),
		),
		ShowConvertedFiles: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "show/hide converted files")),
	}
)

var _ help.KeyMap = ControllerKeyMap{}

type ControllerKeyMap struct {
	Quit     key.Binding
	Activate key.Binding
	Convert  key.Binding
	FullPath key.Binding
}

func (k ControllerKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.Quit,
		k.Activate,
		k.Convert,
		k.FullPath,
	}
}

func (k ControllerKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{k.ShortHelp()}
}

var _ help.KeyMap = FilterKeyMap{}

type FilterKeyMap struct {
	ShowSkippedFiles   key.Binding
	ShowRejectedFiles  key.Binding
	ShowConvertedFiles key.Binding
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
