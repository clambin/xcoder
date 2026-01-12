package refactor

import (
	"testing"

	"github.com/clambin/xcoder/internal/pipeline"
	"github.com/stretchr/testify/assert"
)

func TestFilterState_Show(t *testing.T) {
	tests := []struct {
		name   string
		state  MediaFilterState
		status pipeline.Status
		want   bool
	}{
		{
			name:   "show all",
			state:  MediaFilterState{},
			status: pipeline.Waiting,
			want:   true,
		},
		{
			name:   "hide skipped",
			state:  MediaFilterState{hideSkipped: true},
			status: pipeline.Skipped,
			want:   false,
		},
		{
			name:   "show skipped",
			state:  MediaFilterState{hideSkipped: false},
			status: pipeline.Skipped,
			want:   true,
		},
		{
			name:   "hide rejected",
			state:  MediaFilterState{hideRejected: true},
			status: pipeline.Rejected,
			want:   false,
		},
		{
			name:   "show rejected",
			state:  MediaFilterState{hideRejected: false},
			status: pipeline.Rejected,
			want:   true,
		},
		{
			name:   "hide converted",
			state:  MediaFilterState{hideConverted: true},
			status: pipeline.Converted,
			want:   false,
		},
		{
			name:   "show converted",
			state:  MediaFilterState{hideConverted: false},
			status: pipeline.Converted,
			want:   true,
		},
		{
			name:   "other status (waiting)",
			state:  MediaFilterState{hideSkipped: true, hideRejected: true, hideConverted: true},
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
		state MediaFilterState
		want  string
	}{
		{
			name:  "all show",
			state: MediaFilterState{hideSkipped: false, hideRejected: false, hideConverted: false},
			want:  "",
		},
		{
			name:  "none show",
			state: MediaFilterState{hideSkipped: true, hideRejected: true, hideConverted: true},
			want:  "none",
		},
		{
			name:  "only skipped",
			state: MediaFilterState{hideSkipped: false, hideRejected: true, hideConverted: true},
			want:  "skipped",
		},
		{
			name:  "only rejected",
			state: MediaFilterState{hideSkipped: true, hideRejected: false, hideConverted: true},
			want:  "rejected",
		},
		{
			name:  "only converted",
			state: MediaFilterState{hideSkipped: true, hideRejected: true, hideConverted: false},
			want:  "converted",
		},
		{
			name:  "skipped and rejected",
			state: MediaFilterState{hideSkipped: false, hideRejected: false, hideConverted: true},
			want:  "rejected,skipped",
		},
		{
			name:  "skipped and converted",
			state: MediaFilterState{hideSkipped: false, hideRejected: true, hideConverted: false},
			want:  "converted,skipped",
		},
		{
			name:  "rejected and converted",
			state: MediaFilterState{hideSkipped: true, hideRejected: false, hideConverted: false},
			want:  "converted,rejected",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.state.String())
		})
	}
}
