package ui

import (
	"fmt"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"codeberg.org/clambin/bubbles/table"
	"github.com/charmbracelet/x/exp/golden"
	"github.com/clambin/xcoder/ffmpeg"
	"github.com/clambin/xcoder/internal/transcoder"
	"github.com/stretchr/testify/assert"
)

func TestWorkItemsViewer_View(t *testing.T) {
	var skippedWorkItem, rejectedWorkItem, convertedWorkItem, failedWorkItem, transcodingWorkItem transcoder.WorkItem
	skippedWorkItem.SetStatus(transcoder.StatusSkipped, assert.AnError)
	rejectedWorkItem.SetStatus(transcoder.StatusRejected, assert.AnError)
	convertedWorkItem.SetStatus(transcoder.StatusConverted, nil)
	failedWorkItem.SetStatus(transcoder.StatusFailed, assert.AnError)
	transcodingWorkItem.SetStatus(transcoder.StatusTranscoding, nil)

	workItems := []*transcoder.WorkItem{&skippedWorkItem, &rejectedWorkItem, &convertedWorkItem, &failedWorkItem, &transcodingWorkItem}
	for i, workItem := range workItems {
		workItem.Source.Path = fmt.Sprintf("file_%d", i)
		workItem.Source.VideoStats = ffmpeg.VideoStats{VideoCodec: "h264", Height: 1080, BitRate: 8_000_000}
		workItem.Target.VideoStats = ffmpeg.VideoStats{VideoCodec: "hevc", Height: 1080, BitRate: 4_000_000}
	}

	tests := []struct {
		name string
		mediaFilterState
	}{
		{"no filter", mediaFilterState{}},
		{"no skipped", mediaFilterState{hideSkipped: true}},
		{"no skipped, rejected", mediaFilterState{hideSkipped: true, hideRejected: true}},
		{"no skipped, rejected, converted", mediaFilterState{hideSkipped: true, hideRejected: true, hideConverted: true}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := workItemsViewer{
				FilterTable: table.NewFilterTable().Columns(workItemsColumns),
				styles:      DefaultStyles().MediaViewerItemStyles,
			}.SetSize(100, 10)

			v, _ = v.Update(refreshTableCmd(workItems, tt.mediaFilterState, true)())
			golden.RequireEqual(t, v.View())
		})
	}
}

func TestTranscodeSessionsViewer(t *testing.T) {
	v := transcodeSessionsViewer{}.Width(100)

	session := transcoder.Session{
		WorkItem: &transcoder.WorkItem{},
	}

	msg := transcodeSessionEventMsg(transcoder.SessionEvent{
		Session: &session,
		Type:    transcoder.SessionStartedEvent,
	})

	var cmd tea.Cmd
	v, cmd = v.Update(msg)
	assert.NotNil(t, cmd)
	assert.Len(t, v.sessions, 1)
	assert.Equal(t, 1, lipgloss.Height(v.View()))

	msg = transcodeSessionEventMsg(transcoder.SessionEvent{
		Session: &session,
		Type:    transcoder.SessionStoppedEvent,
	})

	v, cmd = v.Update(msg)
	assert.NotNil(t, cmd)
	assert.Len(t, v.sessions, 0)
	assert.Empty(t, v.View())
}

func TestMediaFilterState_String(t *testing.T) {
	tests := []struct {
		name string
		mediaFilterState
		want string
	}{
		{"empty", mediaFilterState{}, ""},
		{"single", mediaFilterState{hideSkipped: true}, "!skipped"},
		{"multiple", mediaFilterState{hideSkipped: true, hideConverted: true}, "!converted,!skipped"},
		{"all", mediaFilterState{hideSkipped: true, hideConverted: true, hideRejected: true}, "!converted,!rejected,!skipped"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.String())
		})
	}
}

func Test_ltrim(t *testing.T) {
	tests := []struct {
		name  string
		input string
		n     int
		want  string
	}{
		{"empty", "", 10, ""},
		{"short", "012", 10, "012"},
		{"exact", "0123456789", 10, "0123456789"},
		{"long", "012345678912", 10, "…345678912"},
		{"unicode short", "012🤷", 10, "012🤷"},
		{"unicode exact", "012345678🤷", 10, "012345678🤷"},
		{"unicode long", "012345678912🤷", 10, "…45678912🤷"},
		{"no space", "0123", 0, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ltrim(tt.input, tt.n, '…')
			assert.Equal(t, tt.want, got)
		})
	}
}
