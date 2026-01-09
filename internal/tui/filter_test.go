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

func TestFilter(t *testing.T) {
	keyMap := defaultFilterKeyMap()
	var f tea.Model = filter{keyMap: keyMap}

	cmd := f.Init()
	require.NotNil(t, cmd)
	msg := cmd()
	require.IsType(t, filterStateChangedMsg{}, msg)
	assert.Equal(t, filterState{}, filterState(msg.(filterStateChangedMsg)))

	// note: these must be executed in order
	tests := []struct {
		key  string
		want filterState
	}{
		{"s", filterState{true, false, false}},
		{"s", filterState{false, false, false}},
		{"r", filterState{false, true, false}},
		{"c", filterState{false, true, true}},
	}
	for _, tt := range tests {
		var cmd tea.Cmd
		f, cmd = f.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)})
		assert.Equal(t, tt.want, f.(filter).filterState)
		require.NotNil(t, cmd)
		msg := cmd()
		require.IsType(t, filterStateChangedMsg{}, msg)
		assert.Equal(t, tt.want, filterState(msg.(filterStateChangedMsg)))
	}

	// Test unknown key
	_, cmd = f.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})
	assert.Nil(t, cmd)

	// filter has no output
	assert.Equal(t, "", f.View())
}
