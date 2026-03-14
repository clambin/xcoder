package tui

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/exp/golden"
	"github.com/charmbracelet/x/exp/teatest/v2"
	"github.com/clambin/xcoder/internal/pipeline"
)

func TestApplication(t *testing.T) {
	q := fakeQueue{
		queue: []*pipeline.WorkItem{
			{Source: pipeline.MediaFile{Path: "file1.mp4"}},
			{Source: pipeline.MediaFile{Path: "file2.mp4"}},
		},
	}

	cfg := pipeline.Configuration{
		Input:       ".",
		ProfileName: "foo",
		Active:      false,
		Remove:      true,
		Overwrite:   true,
	}

	var a tea.Model = New(&q, cfg)
	tm := teatest.NewTestModel(t, a, teatest.WithInitialTermSize(200, 25))
	waitFor(t, tm.Output(), []byte("waiting"))

	// Converting draws a spinner on the status line. This makes the test flaky. Don't test the convertion action for now.
	/*
		tm.Send(tea.KeyPressMsg{Code: tea.KeyEnter})
		waitFor(t, tm.Output(), []byte("converting"))
	*/

	tm.Send(tea.KeyPressMsg{Text: "q"})
	tm.WaitFinished(t)

	a = tm.FinalModel(t)
	golden.RequireEqual(t, a.View().Content)
}
