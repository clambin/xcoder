package tui

import (
	"maps"
	"slices"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/clambin/xcoder/internal/pipeline"
)

// filterStateChangedMsg is a BubbleTea message that indicates that the media filter changed state.
// Its value is the new filterState.
type filterStateChangedMsg filterState

// filterState holds the current state of the media filter. It determines which media files should be shown/hidden.
type filterState struct {
	hideSkipped   bool
	hideRejected  bool
	hideConverted bool
}

// Show returns true if the given item should be shown
func (s filterState) Show(item *pipeline.WorkItem) bool {
	switch item.WorkStatus().Status {
	case pipeline.Rejected:
		return !s.hideRejected
	case pipeline.Converted:
		return !s.hideConverted
	case pipeline.Skipped:
		return !s.hideSkipped
	default:
		return true
	}
}

// String returns a string representation of the filter state
func (s filterState) String() string {
	on := map[string]struct{}{
		"skipped":   {},
		"rejected":  {},
		"converted": {},
	}
	if s.hideSkipped {
		delete(on, "skipped")
	}
	if s.hideRejected {
		delete(on, "rejected")
	}
	if s.hideConverted {
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

// filter determines which media files should be shown/hidden
type filter struct {
	keyMap      FilterKeyMap
	filterState filterState
}

func (f filter) Update(msg tea.Msg) (filter, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, f.keyMap.ShowSkippedFiles):
			f.filterState.hideSkipped = !f.filterState.hideSkipped
			return f, func() tea.Msg { return filterStateChangedMsg(f.filterState) }
		case key.Matches(msg, f.keyMap.ShowRejectedFiles):
			f.filterState.hideRejected = !f.filterState.hideRejected
			return f, func() tea.Msg { return filterStateChangedMsg(f.filterState) }
		case key.Matches(msg, f.keyMap.ShowConvertedFiles):
			f.filterState.hideConverted = !f.filterState.hideConverted
			return f, func() tea.Msg { return filterStateChangedMsg(f.filterState) }
		}
	}
	return f, nil
}
