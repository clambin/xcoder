package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/clambin/xcoder/internal/pipeline"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFilterState_Show(t *testing.T) {
	tests := []struct {
		name   string
		state  filterState
		status pipeline.Status
		want   bool
	}{
		{
			name:   "show all",
			state:  filterState{},
			status: pipeline.Waiting,
			want:   true,
		},
		{
			name:   "hide skipped",
			state:  filterState{hideSkipped: true},
			status: pipeline.Skipped,
			want:   false,
		},
		{
			name:   "show skipped",
			state:  filterState{hideSkipped: false},
			status: pipeline.Skipped,
			want:   true,
		},
		{
			name:   "hide rejected",
			state:  filterState{hideRejected: true},
			status: pipeline.Rejected,
			want:   false,
		},
		{
			name:   "show rejected",
			state:  filterState{hideRejected: false},
			status: pipeline.Rejected,
			want:   true,
		},
		{
			name:   "hide converted",
			state:  filterState{hideConverted: true},
			status: pipeline.Converted,
			want:   false,
		},
		{
			name:   "show converted",
			state:  filterState{hideConverted: false},
			status: pipeline.Converted,
			want:   true,
		},
		{
			name:   "other status (waiting)",
			state:  filterState{hideSkipped: true, hideRejected: true, hideConverted: true},
			status: pipeline.Waiting,
			want:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var item pipeline.WorkItem
			item.SetWorkStatus(pipeline.WorkStatus{Status: tt.status})
			assert.Equal(t, tt.want, tt.state.Show(&item))
		})
	}
}

func TestFilterState_String(t *testing.T) {
	tests := []struct {
		name  string
		state filterState
		want  string
	}{
		{
			name:  "show all",
			state: filterState{},
			want:  "",
		},
		{
			name:  "hide all",
			state: filterState{hideSkipped: true, hideRejected: true, hideConverted: true},
			want:  "none",
		},
		{
			name:  "hide skipped",
			state: filterState{hideSkipped: true},
			want:  "converted,rejected",
		},
		{
			name:  "hide rejected",
			state: filterState{hideRejected: true},
			want:  "converted,skipped",
		},
		{
			name:  "hide converted",
			state: filterState{hideConverted: true},
			want:  "rejected,skipped",
		},
		{
			name:  "show only skipped",
			state: filterState{hideRejected: true, hideConverted: true},
			want:  "skipped",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.state.String())
		})
	}
}

func TestFilter_Update(t *testing.T) {
	keyMap := defaultFilterKeyMap()
	f := filter{
		keyMap: keyMap,
	}

	// Test ShowSkippedFiles
	var cmd tea.Cmd
	f, cmd = f.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")})
	assert.True(t, f.filterState.hideSkipped)
	checkFilterState(t, f.filterState, cmd)

	// Toggle back
	f, cmd = f.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")})
	assert.False(t, f.filterState.hideSkipped)
	checkFilterState(t, f.filterState, cmd)

	// Test ShowRejectedFiles
	f, cmd = f.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	assert.True(t, f.filterState.hideRejected)
	checkFilterState(t, f.filterState, cmd)

	// Test ShowConvertedFiles
	f, cmd = f.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("c")})
	assert.True(t, f.filterState.hideConverted)
	checkFilterState(t, f.filterState, cmd)

	// Test unknown key
	_, cmd = f.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})
	assert.Nil(t, cmd)
}

func checkFilterState(t *testing.T, want filterState, cmd tea.Cmd) {
	t.Helper()
	require.NotNil(t, cmd)
	msg := cmd()
	assert.IsType(t, filterStateChangedMsg{}, msg)
	assert.Equal(t, want, filterState(msg.(filterStateChangedMsg)))
}
