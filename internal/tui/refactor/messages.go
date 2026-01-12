package refactor

// MediaFilterChangedMsg is sent when the MediaFilter changes
type MediaFilterChangedMsg MediaFilterState

// RefreshUIMsg is sent to update all UI components
type RefreshUIMsg struct{}

// LogViewerClosedMsg indicates that the log viewer should be closed
type LogViewerClosedMsg struct{}
