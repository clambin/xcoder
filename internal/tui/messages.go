package tui

// MediaFilterChangedMsg is sent when the MediaFilter changes
type MediaFilterChangedMsg MediaFilterState

// RefreshUIMsg updates all UI components
type RefreshUIMsg struct{}

// AutoRefreshUIMsg is the tick msg to automatically refresh all UI components
type AutoRefreshUIMsg struct{}

// LogViewerClosedMsg indicates that the log viewer should be closed
type LogViewerClosedMsg struct{}
