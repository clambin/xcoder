package tui

import (
	"maps"
	"slices"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/clambin/xcoder/internal/pipeline"
)

// MediaFilterState holds the current state of the MediaFilter. It determines which media files should be shown/hidden.
type MediaFilterState struct {
	hideSkipped   bool
	hideRejected  bool
	hideConverted bool
}

// Show returns true if the given item should be shown
func (s MediaFilterState) Show(item *pipeline.WorkItem) bool {
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

// String returns a string representation of the MediaFilterState
func (s MediaFilterState) String() string {
	on := map[string]struct{}{"skipped": {}, "rejected": {}, "converted": {}}
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

// MediaFilter determines which media files should be shown/hidden
type MediaFilter struct {
	KeyMap           MediaFilterKeyMap
	mediaFilterState MediaFilterState
}

func (f *MediaFilter) Init() tea.Cmd {
	return func() tea.Msg { return MediaFilterChangedMsg(f.mediaFilterState) }
}

func (f *MediaFilter) Update(msg tea.Msg) tea.Cmd {
	var action bool
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, f.KeyMap.ShowSkippedFiles):
			f.mediaFilterState.hideSkipped = !f.mediaFilterState.hideSkipped
			action = true
		case key.Matches(msg, f.KeyMap.ShowRejectedFiles):
			f.mediaFilterState.hideRejected = !f.mediaFilterState.hideRejected
			action = true
		case key.Matches(msg, f.KeyMap.ShowConvertedFiles):
			f.mediaFilterState.hideConverted = !f.mediaFilterState.hideConverted
			action = true
		}
	}
	if !action {
		return nil
	}
	return tea.Batch(
		func() tea.Msg { return MediaFilterChangedMsg(f.mediaFilterState) },
		func() tea.Msg { return RefreshUIMsg{} },
	)
}

func (f *MediaFilter) View() string {
	// not used: controller renders mediaFilter state directly from MediaFilterChangedMsg
	return ""
}
