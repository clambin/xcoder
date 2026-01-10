package tui

import (
	"iter"
	"path/filepath"
	"time"

	"codeberg.org/clambin/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/clambin/xcoder/internal/pipeline"
)

// paneSizeMsg resizes the pane to the given width and height.
type paneSizeMsg struct {
	Width  int
	Height int
}

func paneSizeCmd(width, height int) tea.Cmd {
	return cmd(paneSizeMsg{Width: width, Height: height})
}

type showLogsMsg struct {
	on bool
}

func showLogsCmd(on bool) tea.Cmd {
	return cmd(showLogsMsg{on})
}

type showFullPathMsg struct {
	on bool
}

func showFullPathCmd(on bool) tea.Cmd {
	return cmd(showFullPathMsg{on})
}

type textFilterStateChangMsg struct {
	on bool
}

func textFilterStateChangCmd(on bool) tea.Cmd {
	return cmd(textFilterStateChangMsg{on})
}

// autoRefreshMsg refreshes the screen at a regular interval.
type autoRefreshMsg struct{}

func autoRefreshCmd() func(_ time.Time) tea.Msg {
	return func(_ time.Time) tea.Msg {
		return autoRefreshMsg{}
	}
}

// refreshTableMsg refreshes the table pane.
type refreshTableMsg struct{}

func refreshTableCmd() tea.Cmd {
	return cmd(refreshTableMsg{})
}

// mediaFilterActivateMsg tells mediaFilter if it's active or not.
type mediaFilterActivateMsg struct {
	active bool
}

func mediaFilterActivateCmd(active bool) tea.Cmd {
	return cmd(mediaFilterActivateMsg{active})
}

// mediaFilterChangedMsg indicates that the media mediaFilter changed state.
// Its value is the new mediaFilterState.
type mediaFilterChangedMsg mediaFilterState

// cmd is a helper function to create a tea.Cmd that returns a tea.Msg.
func cmd(msg tea.Msg) tea.Cmd {
	return func() tea.Msg { return msg }
}

// setTitleCmd sets the title of the table's frame.
func setTitleCmd(f mediaFilterState, s lipgloss.Style) tea.Cmd {
	return func() tea.Msg {
		args := f.String()
		if args != "" {
			args = " (" + s.Render(args) + ")"
		}
		return table.SetTitleMsg{Title: "media files" + args}
	}
}

// loadTableCmd builds the table with the current Queue state and issues a command to load it in the table.
func loadTableCmd(items iter.Seq[*pipeline.WorkItem], f mediaFilterState, showFullPath bool) tea.Cmd {
	return func() tea.Msg {
		var rows []table.Row
		for item := range items {
			if !f.Show(item) {
				continue
			}
			source := item.Source.Path
			if !showFullPath {
				source = filepath.Base(source)
			}
			workStatus := item.WorkStatus()
			var errString string
			if workStatus.Err != nil {
				errString = workStatus.Err.Error()
			}
			rows = append(rows, table.Row{
				source,
				item.SourceVideoStats().String(),
				item.TargetVideoStats().String(),
				workStatus.Status.String(),
				item.CompletedFormatted(),
				item.RemainingFormatted(),
				errString,
				table.UserData{Data: item},
			})
		}
		return table.SetRowsMsg{Rows: rows}
	}
}
