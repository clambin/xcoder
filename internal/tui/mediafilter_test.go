package tui

import (
	"testing"

	"github.com/clambin/xcoder/internal/pipeline"
	"github.com/stretchr/testify/assert"
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

func TestFilterState_String(t *testing.T) {
	tests := []struct {
		name  string
		state mediaFilterState
		want  string
	}{
		{
			name:  "all show",
			state: mediaFilterState{hideSkipped: false, hideRejected: false, hideConverted: false},
			want:  "",
		},
		{
			name:  "none show",
			state: mediaFilterState{hideSkipped: true, hideRejected: true, hideConverted: true},
			want:  "none",
		},
		{
			name:  "only skipped",
			state: mediaFilterState{hideSkipped: false, hideRejected: true, hideConverted: true},
			want:  "skipped",
		},
		{
			name:  "only rejected",
			state: mediaFilterState{hideSkipped: true, hideRejected: false, hideConverted: true},
			want:  "rejected",
		},
		{
			name:  "only converted",
			state: mediaFilterState{hideSkipped: true, hideRejected: true, hideConverted: false},
			want:  "converted",
		},
		{
			name:  "skipped and rejected",
			state: mediaFilterState{hideSkipped: false, hideRejected: false, hideConverted: true},
			want:  "rejected,skipped",
		},
		{
			name:  "skipped and converted",
			state: mediaFilterState{hideSkipped: false, hideRejected: true, hideConverted: false},
			want:  "converted,skipped",
		},
		{
			name:  "rejected and converted",
			state: mediaFilterState{hideSkipped: true, hideRejected: false, hideConverted: false},
			want:  "converted,rejected",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.state.String())
		})
	}
}
