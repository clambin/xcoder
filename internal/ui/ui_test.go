package ui

import (
	"bytes"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/exp/golden"
	"github.com/charmbracelet/x/exp/teatest/v2"
	"github.com/clambin/xcoder/ffmpeg"
	"github.com/clambin/xcoder/internal/transcoder"
	"github.com/stretchr/testify/assert"
)

func TestApplication(t *testing.T) {
	tests := []struct {
		name    string
		keys    []tea.Key
		waitFor string
	}{
		{"main", nil, ""},
		{"logs", []tea.Key{{Text: "l"}}, "Autoscroll:"},
		{"help", []tea.Key{{Text: "?"}}, "esc close logs"},
		{"no skip", []tea.Key{{Text: "s"}}, "[!skipped]"},
		{"no reject", []tea.Key{{Text: "r"}}, "[!rejected]"},
		{"no convert", []tea.Key{{Text: "c"}}, "[!converted]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buff bytes.Buffer
			for i := range 3 {
				_, _ = fmt.Fprintf(&buff, "line %d\n", i+1)
			}
			q := generateWorkItems()
			var a tea.Model = New(q, &fakeTranscoder{}, "test", &buff, DefaultKeyMap(), DefaultStyles())
			tm := teatest.NewTestModel(t, a, teatest.WithInitialTermSize(120, 10))
			teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
				return bytes.Contains(bts, []byte("hevc"))
			})

			if len(tt.keys) > 0 {
				for _, key := range tt.keys {
					tm.Send(tea.KeyPressMsg(key))
				}
				teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
					return bytes.Contains(bts, []byte(tt.waitFor))
				}, teatest.WithDuration(5*time.Second), teatest.WithCheckInterval(50*time.Millisecond))
			}

			tm.Send(tea.KeyPressMsg(tea.Key{Text: "q"}))

			golden.RequireEqual(t, tm.FinalModel(t).View().Content)
		})
	}
}

func generateWorkItems() *transcoder.WorkItems {
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
	var q transcoder.WorkItems
	q.Add(workItems...)
	return &q
}

var _ Transcoder = (*fakeTranscoder)(nil)

type fakeTranscoder struct {
	active atomic.Bool
	count  int
}

func (f *fakeTranscoder) SessionCount() int {
	return f.count
}

func (f *fakeTranscoder) Active() bool {
	return f.active.Load()
}

func (f *fakeTranscoder) SetActive(active bool) {
	f.active.Store(active)
}

func (f *fakeTranscoder) Subscribe() <-chan transcoder.SessionEvent {
	return nil
}

func (f *fakeTranscoder) OverwriteTarget() bool {
	return true
}

func (f *fakeTranscoder) RemoveSource() bool {
	return true
}
