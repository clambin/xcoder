package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/clambin/xcoder/internal/pipeline"
	"github.com/stretchr/testify/assert"
)

func TestStatusLine_BatchStatus(t *testing.T) {
	const expectedWidth = 30
	var q fakeQueue
	s := newStatusLine(&q, StatusStyles{})
	s.SetSize(expectedWidth, 1)

	tests := []struct {
		msg    tea.Msg
		status bool
		want   string
	}{
		{s.Init()(), false, "       Batch processing: OFF  "},
		{nil, true, "       Batch processing: ON   "},
		{s.spinner.Tick(), true, "       Batch processing:      "},
		{s.spinner.Tick(), false, "       Batch processing: OFF  "},
		{s.spinner.Tick(), false, "       Batch processing: OFF  "},
	}

	for _, tt := range tests {
		q.active.Store(tt.status)
		s.Update(tt.msg)
		got := s.View()
		assert.Equal(t, tt.want, got)
		assert.Len(t, got, expectedWidth)
	}
}

func TestStatusLine_Converting(t *testing.T) {
	const expectedWidth = 30
	q := fakeQueue{
		queue: []*pipeline.WorkItem{
			{Source: pipeline.MediaFile{Path: "file1.mp4"}, Target: pipeline.MediaFile{Path: "file1.hevc.mkv"}},
			{Source: pipeline.MediaFile{Path: "file2.mp4"}, Target: pipeline.MediaFile{Path: "file2.hevc.mkv"}},
			{Source: pipeline.MediaFile{Path: "file3.mp4"}, Target: pipeline.MediaFile{Path: "file3.hevc.mkv"}},
		},
	}
	q.queue[0].SetWorkStatus(pipeline.WorkStatus{Status: pipeline.Converting})
	q.queue[1].SetWorkStatus(pipeline.WorkStatus{Status: pipeline.Converting})
	s := newStatusLine(&q, StatusStyles{})
	s.SetSize(expectedWidth, 1)
	s.Update(s.Init()())

	assert.Equal(t, "  Converting 2 file(s) ... ⣽ Batch processing: OFF  ", s.View())
	s.Update(s.spinner.Tick())
	assert.Equal(t, "  Converting 2 file(s) ... ⣻ Batch processing: OFF  ", s.View())
}
