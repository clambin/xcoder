package tui

import (
	"testing"

	"codeberg.org/clambin/bubbles/table"
	"github.com/charmbracelet/lipgloss"
	"github.com/clambin/xcoder/ffmpeg"
	"github.com/clambin/xcoder/internal/pipeline"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_setTitleCmd(t *testing.T) {
	tests := []struct {
		name  string
		state mediaFilterState
		want  string
	}{
		{
			name:  "show all",
			state: mediaFilterState{},
			want:  "",
		},
		{
			name:  "hide all",
			state: mediaFilterState{hideSkipped: true, hideRejected: true, hideConverted: true},
			want:  "none",
		},
		{
			name:  "hide skipped",
			state: mediaFilterState{hideSkipped: true},
			want:  "converted,rejected",
		},
		{
			name:  "hide rejected",
			state: mediaFilterState{hideRejected: true},
			want:  "converted,skipped",
		},
		{
			name:  "hide converted",
			state: mediaFilterState{hideConverted: true},
			want:  "rejected,skipped",
		},
		{
			name:  "show only skipped",
			state: mediaFilterState{hideRejected: true, hideConverted: true},
			want:  "skipped",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := setTitleCmd(tt.state, lipgloss.NewStyle())()
			require.IsType(t, table.SetTitleMsg{}, msg)
			title := msg.(table.SetTitleMsg).Title
			want := "media files"
			if tt.want != "" {
				want += " (" + tt.want + ")"
			}
			assert.Equal(t, want, want, title)
		})
	}

}

func Test_loadTableCmd(t *testing.T) {
	inVideoStats := ffmpeg.VideoStats{VideoCodec: "h264", Height: 1080, BitRate: 8_000_000}
	outVideoStats := ffmpeg.VideoStats{VideoCodec: "hevc", Height: 1080, BitRate: 4_000_000}
	workItems := []*pipeline.WorkItem{
		{
			Source: pipeline.MediaFile{Path: "/foo/video1.mkv", VideoStats: inVideoStats},
			Target: pipeline.MediaFile{VideoStats: outVideoStats},
		},
		{
			Source: pipeline.MediaFile{Path: "/foo/video2.mkv", VideoStats: inVideoStats},
			Target: pipeline.MediaFile{VideoStats: outVideoStats},
		},
		{
			Source: pipeline.MediaFile{Path: "/foo/video3.mkv", VideoStats: inVideoStats},
			Target: pipeline.MediaFile{VideoStats: outVideoStats},
		},
	}
	workItems[0].SetWorkStatus(pipeline.WorkStatus{Status: pipeline.Skipped})
	workItems[2].SetWorkStatus(pipeline.WorkStatus{Status: pipeline.Failed, Err: assert.AnError})

	items := func(yield func(item *pipeline.WorkItem) bool) {
		for _, item := range workItems {
			if !yield(item) {
				return
			}
		}
	}
	f := mediaFilterState{hideSkipped: true}
	msg := loadTableCmd(items, f, false)()
	require.IsType(t, table.SetRowsMsg{}, msg)
	want := []table.Row{
		{"video2.mkv", "h264/1080/8.00 mbps", "hevc/1080/4.00 mbps", "waiting", "", "", "", table.UserData{Data: workItems[1]}},
		{"video3.mkv", "h264/1080/8.00 mbps", "hevc/1080/4.00 mbps", "failed", "", "", "assert.AnError general error for testing", table.UserData{Data: workItems[2]}},
	}
	assert.Equal(t, want, msg.(table.SetRowsMsg).Rows)
}
