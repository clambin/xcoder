package tui

import (
	"maps"
	"slices"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/clambin/xcoder/internal/pipeline"
)

type filter struct {
	keyMap        FilterKeyMap
	hideSkipped   bool
	hideRejected  bool
	hideConverted bool
}

func (f filter) Update(msg tea.Msg) (filter, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, f.keyMap.ShowSkippedFiles):
			f.hideSkipped = !f.hideSkipped
			return f, refreshTableCmd()
		case key.Matches(msg, f.keyMap.ShowRejectedFiles):
			f.hideRejected = !f.hideRejected
			return f, refreshTableCmd()
		case key.Matches(msg, f.keyMap.ShowConvertedFiles):
			f.hideConverted = !f.hideConverted
			return f, refreshTableCmd()
		}
	}
	return f, nil
}

func (f filter) Show(item *pipeline.WorkItem) bool {
	switch item.WorkStatus().Status {
	case pipeline.Rejected:
		return !f.hideRejected
	case pipeline.Converted:
		return !f.hideConverted
	case pipeline.Skipped:
		return !f.hideSkipped
	default:
		return true
	}
}

func (f filter) String() string {
	on := map[string]struct{}{
		"skipped":   {},
		"rejected":  {},
		"converted": {},
	}
	if f.hideSkipped {
		delete(on, "skipped")
	}
	if f.hideRejected {
		delete(on, "rejected")
	}
	if f.hideConverted {
		delete(on, "converted")
	}
	if len(on) == 3 {
		return ""
	}
	if len(on) == 0 {
		return "none"
	}
	onString := slices.Collect(maps.Keys(on))
	slices.Sort(onString)
	return strings.Join(onString, ",")
}
