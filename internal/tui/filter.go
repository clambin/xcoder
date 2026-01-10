package tui

import (
	"maps"
	"slices"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/clambin/xcoder/internal/pipeline"
)

// mediaFilterState holds the current state of the media mediaFilter. It determines which media files should be shown/hidden.
type mediaFilterState struct {
	hideSkipped   bool
	hideRejected  bool
	hideConverted bool
}

// Show returns true if the given item should be shown
func (s mediaFilterState) Show(item *pipeline.WorkItem) bool {
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

// String returns a string representation of the mediaFilter state
func (s mediaFilterState) String() string {
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

var _ tea.Model = mediaFilter{}

// mediaFilter determines which media files should be shown/hidden
type mediaFilter struct {
	keyMap           FilterKeyMap
	mediaFilterState mediaFilterState
	active           bool
}

func (f mediaFilter) Init() tea.Cmd {
	return func() tea.Msg { return mediaFilterChangedMsg(f.mediaFilterState) }
}

func (f mediaFilter) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case mediaFilterActivateMsg:
		f.active = msg.active
	case tea.KeyMsg:
		if !f.active {
			break
		}
		switch {
		case key.Matches(msg, f.keyMap.ShowSkippedFiles):
			f.mediaFilterState.hideSkipped = !f.mediaFilterState.hideSkipped
			return f, func() tea.Msg { return mediaFilterChangedMsg(f.mediaFilterState) }
		case key.Matches(msg, f.keyMap.ShowRejectedFiles):
			f.mediaFilterState.hideRejected = !f.mediaFilterState.hideRejected
			return f, func() tea.Msg { return mediaFilterChangedMsg(f.mediaFilterState) }
		case key.Matches(msg, f.keyMap.ShowConvertedFiles):
			f.mediaFilterState.hideConverted = !f.mediaFilterState.hideConverted
			return f, func() tea.Msg { return mediaFilterChangedMsg(f.mediaFilterState) }
		}
	}
	return f, nil
}

func (f mediaFilter) View() string {
	// not used: controller renders mediaFilter state directly
	return ""
}
