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
		state  mediaFilterState
		status pipeline.Status
		want   bool
	}{
		{
			name:   "show all",
			state:  mediaFilterState{},
			status: pipeline.Waiting,
			want:   true,
		},
		{
			name:   "hide skipped",
			state:  mediaFilterState{hideSkipped: true},
			status: pipeline.Skipped,
			want:   false,
		},
		{
			name:   "show skipped",
			state:  mediaFilterState{hideSkipped: false},
			status: pipeline.Skipped,
			want:   true,
		},
		{
			name:   "hide rejected",
			state:  mediaFilterState{hideRejected: true},
			status: pipeline.Rejected,
			want:   false,
		},
		{
			name:   "show rejected",
			state:  mediaFilterState{hideRejected: false},
			status: pipeline.Rejected,
			want:   true,
		},
		{
			name:   "hide converted",
			state:  mediaFilterState{hideConverted: true},
			status: pipeline.Converted,
			want:   false,
		},
		{
			name:   "show converted",
			state:  mediaFilterState{hideConverted: false},
			status: pipeline.Converted,
			want:   true,
		},
		{
			name:   "other status (waiting)",
			state:  mediaFilterState{hideSkipped: true, hideRejected: true, hideConverted: true},
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
	var f tea.Model = mediaFilter{keyMap: keyMap}

	cmd := f.Init()
	require.NotNil(t, cmd)
	msg := cmd()
	require.IsType(t, mediaFilterChangedMsg{}, msg)
	assert.Equal(t, mediaFilterState{}, mediaFilterState(msg.(mediaFilterChangedMsg)))

	// active the mediaFilter
	f, _ = f.Update(mediaFilterActivateMsg{true})

	// note: these must be executed in order
	tests := []struct {
		key  string
		want mediaFilterState
	}{
		{"s", mediaFilterState{true, false, false}},
		{"s", mediaFilterState{false, false, false}},
		{"r", mediaFilterState{false, true, false}},
		{"c", mediaFilterState{false, true, true}},
	}
	for _, tt := range tests {
		var cmd tea.Cmd
		f, cmd = f.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)})
		assert.Equal(t, tt.want, f.(mediaFilter).mediaFilterState)
		require.NotNil(t, cmd)
		msg := cmd()
		require.IsType(t, mediaFilterChangedMsg{}, msg)
		assert.Equal(t, tt.want, mediaFilterState(msg.(mediaFilterChangedMsg)))
	}

	// Test unknown key
	_, cmd = f.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})
	assert.Nil(t, cmd)

	// mediaFilter has no output
	assert.Equal(t, "", f.View())
}
