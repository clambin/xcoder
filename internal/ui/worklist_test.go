package ui

import (
	"github.com/clambin/go-common/set"
	"github.com/clambin/videoConvertor/internal/profile"
	"github.com/clambin/videoConvertor/internal/worklist"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_workListViewer(t *testing.T) {
	var list worklist.WorkList
	list.Add("A").SetStatus(worklist.Skipped, profile.ErrSourceInTargetCodec)
	list.Add("B").SetStatus(worklist.Rejected, profile.ErrSourceRejected{Reason: "bitrate too low"})
	list.Add("C").SetStatus(worklist.Inspected, nil)
	list.Add("D").SetStatus(worklist.Converted, nil)

	tests := []struct {
		name      string
		filters   []worklist.WorkStatus
		wantCount int
		wantFirst string
	}{
		{
			name:      "no filters",
			wantCount: 1 + len(list.List()),
			wantFirst: "A",
		},
		{
			name:      "filter skipped",
			filters:   []worklist.WorkStatus{worklist.Skipped},
			wantCount: 1 + len(list.List()) - 1,
			wantFirst: "B",
		},
		{
			name:      "filter rejected",
			filters:   []worklist.WorkStatus{worklist.Rejected},
			wantCount: 1 + len(list.List()) - 1,
			wantFirst: "A",
		},
		{
			name:      "filter skipped & rejected",
			filters:   []worklist.WorkStatus{worklist.Skipped, worklist.Rejected},
			wantCount: 1 + len(list.List()) - 2,
			wantFirst: "C",
		},
		{
			name:      "all filter skipped & rejected",
			filters:   []worklist.WorkStatus{worklist.Skipped, worklist.Rejected, worklist.Inspected, worklist.Converted},
			wantCount: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := newWorkListViewer(&list)
			v.DataSource.(*workItems).toggle(tt.filters...)
			v.refresh()

			assert.Equal(t, tt.wantCount, v.GetRowCount())
			if tt.wantFirst != "" {
				require.Greater(t, v.GetRowCount(), 1)
				assert.Equal(t, tt.wantFirst, v.GetCell(1, 0).Text)
			}
		})
	}

}

func Test_workItems_title(t *testing.T) {
	tests := []struct {
		name          string
		filters       []worklist.WorkStatus
		itemCount     int
		rowCount      int
		expectedTitle string
	}{
		{
			name:          "No filter, multiple items",
			itemCount:     10,
			rowCount:      5,
			expectedTitle: " files [5] ",
		},
		{
			name:          "With filter, multiple items",
			filters:       []worklist.WorkStatus{worklist.Skipped},
			itemCount:     10,
			rowCount:      5,
			expectedTitle: " files (filtered: skipped)[5/10] ",
		},
		{
			name:          "With filter, single item",
			filters:       []worklist.WorkStatus{worklist.Skipped, worklist.Rejected},
			itemCount:     1,
			rowCount:      1,
			expectedTitle: " files (filtered: rejected, skipped)[1/1] ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ds := workItems{filters: filters{statuses: set.Set[worklist.WorkStatus]{}}}
			ds.filters.toggle(tt.filters...)
			assert.Equal(t, tt.expectedTitle, ds.title(tt.itemCount, tt.rowCount))
		})
	}
}
