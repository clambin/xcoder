package tui

import "codeberg.org/clambin/bubbles/table"

// mediaFilterChangedMsg is sent when the mediaFilter changes
type mediaFilterChangedMsg mediaFilterState

// refreshUIMsg updates all UI components
type refreshUIMsg struct{}

// autoRefreshUIMsg is the tick msg to automatically refresh all UI components
type autoRefreshUIMsg struct{}

// logViewerClosedMsg indicates that the log viewer should be closed
type logViewerClosedMsg struct{}

// setRowsMsg updates the rows in the table
type setRowsMsg struct {
	rows []table.Row
}
