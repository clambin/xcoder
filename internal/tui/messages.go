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
	return func() tea.Msg {
		return refreshTableMsg{}
	}
}

// setPaneMsg sets the active pane in the UI
type setPaneMsg activePane

func setPaneCmd(pane activePane) tea.Cmd {
	return func() tea.Msg {
		return setPaneMsg(pane)
	}
}

// setTitleCmd sets the title of the table's frame.
func setTitleCmd(f filterState, s lipgloss.Style) tea.Cmd {
	return func() tea.Msg {
		args := f.String()
		if args != "" {
			args = " (" + s.Render(args) + ")"
		}
		return table.SetTitleMsg{Title: "media files" + args}
	}
}

// loadTableCmd builds the table with the current Queue state and issues a command to load it in the table.
func loadTableCmd(items iter.Seq[*pipeline.WorkItem], f filterState, showFullPath bool) tea.Cmd {
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
