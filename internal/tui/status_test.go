package tui

import (
	"testing"
	"unicode/utf8"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/clambin/xcoder/internal/pipeline"
	"github.com/stretchr/testify/assert"
)

func TestStatusLine_BatchStatus(t *testing.T) {
	const expectedWidth = 30
	var q fakeQueue
	s := newStatusLine(&q, StatusStyles{}).SetSize(expectedWidth, 1)

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
		s, _ = s.Update(tt.msg)
		got := s.View()
		assert.Equal(t, tt.want, got)
		assert.Len(t, got, expectedWidth)
	}
}

func TestStatusLine_Converting(t *testing.T) {
	const expectedWidth = 54
	q := fakeQueue{
		queue: []*pipeline.WorkItem{
			{Source: pipeline.MediaFile{Path: "file1.mp4"}, Target: pipeline.MediaFile{Path: "file1.hevc.mkv"}},
			{Source: pipeline.MediaFile{Path: "file2.mp4"}, Target: pipeline.MediaFile{Path: "file2.hevc.mkv"}},
			{Source: pipeline.MediaFile{Path: "file3.mp4"}, Target: pipeline.MediaFile{Path: "file3.hevc.mkv"}},
		},
	}
	q.queue[0].SetWorkStatus(pipeline.WorkStatus{Status: pipeline.Converting})
	q.queue[1].SetWorkStatus(pipeline.WorkStatus{Status: pipeline.Converting})
	s := newStatusLine(&q, StatusStyles{}, spinner.WithSpinner(spinner.Dot)).SetSize(expectedWidth, 1)
	q.active.Store(true)

	v := s.View()
	assert.Equal(t, expectedWidth, utf8.RuneCountInString(ansi.Strip(v)))
	assert.Equal(t, "  Converting 2 file(s) ... ⣾   Batch processing:      ", v)
	s, _ = s.Update(s.spinner.Tick())
	v = s.View()
	assert.Equal(t, expectedWidth, utf8.RuneCountInString(ansi.Strip(v)))
	assert.Equal(t, "  Converting 2 file(s) ... ⣽   Batch processing: ON   ", v)
}
