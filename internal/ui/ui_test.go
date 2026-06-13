package ui

import (
	"bytes"
	"fmt"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/exp/golden"
	"github.com/charmbracelet/x/exp/teatest/v2"
	"github.com/clambin/xcoder/ffmpeg"
	"github.com/clambin/xcoder/internal/transcoder"
	"github.com/stretchr/testify/assert"
)

func TestApplication(t *testing.T) {
	var buff bytes.Buffer
	for i := range 3 {
		buff.WriteString(fmt.Sprintf("line %d", i+1))
	}
	q := generateWorkItems()
	var a tea.Model = New(q, &fakeTranscoder{}, "test", &buff, DefaultKeyMap(), DefaultStyles())
	tm := teatest.NewTestModel(t, a)
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return bytes.Contains(bts, []byte("hevc"))
	})

	tm.Send(tea.QuitMsg{})

	golden.RequireEqual(t, tm.FinalModel(t).View().Content)
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
	active bool
	count  int
}

func (f *fakeTranscoder) SessionCount() int {
	return f.count
}

func (f *fakeTranscoder) Active() bool {
	return f.active
}

func (f *fakeTranscoder) SetActive(_ bool) {
	panic("implement me")
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
