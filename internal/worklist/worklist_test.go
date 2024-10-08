package worklist

import (
	"github.com/clambin/videoConvertor/internal/ffmpeg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestWorkList(t *testing.T) {
	var l WorkList

	// default WorkList is empty and inactive
	assert.Empty(t, l.List())
	assert.False(t, l.Active())

	// manually queued items are returned even if the WorkList is inactive
	l.Queue(&WorkItem{Source: "foo"})
	i := l.NextToConvert()
	assert.Equal(t, "foo", i.Source)
	status, err := i.Status()
	assert.Equal(t, Converting, status)
	assert.NoError(t, err)

	// automatically added items are not returned if the WorkList is inactive
	l.Add("foo").status = Inspected
	i = l.NextToConvert()
	assert.Nil(t, i)

	// automatically added items are not returned if the WorkList is active
	l.SetActive(true)
	i = l.NextToConvert()
	require.NotNil(t, i)
	status, err = i.Status()
	assert.Equal(t, Converting, status)
	assert.NoError(t, err)
}

func TestWorkList_Active(t *testing.T) {
	var l WorkList
	assert.False(t, l.Active())
	l.SetActive(true)
	assert.True(t, l.Active())
	l.ToggleActive()
	assert.False(t, l.Active())
}

func TestWorkItem_Stats(t *testing.T) {
	stats := ffmpeg.NewVideoStats("h264", 1080, 5_000_000)
	var item WorkItem
	item.AddSourceStats(stats)
	assert.Equal(t, stats, item.SourceVideoStats())
	item.AddTargetStats(stats)
	assert.Equal(t, stats, item.TargetVideoStats())
}

func TestWorkStatus_String(t *testing.T) {
	for val, label := range workStatusToString {
		assert.Equal(t, label, val.String())
	}
	assert.Equal(t, "unknown", WorkStatus(-1).String())
}

func TestProgress(t *testing.T) {
	tests := []struct {
		name          string
		duration      time.Duration
		progress      ffmpeg.Progress
		prevSpeed     float64
		wantCompleted float64
		wantRemaining time.Duration
	}{
		{
			name:          "not initialized",
			duration:      time.Hour,
			wantCompleted: 0,
			wantRemaining: -1,
		},
		{
			name:          "half done",
			duration:      time.Hour,
			progress:      ffmpeg.Progress{Speed: 1, Converted: 30 * time.Minute},
			wantCompleted: .5,
			wantRemaining: 30 * time.Minute,
		},
		{
			name:          "speed matters",
			duration:      time.Hour,
			progress:      ffmpeg.Progress{Speed: 2, Converted: 30 * time.Minute},
			wantCompleted: .5,
			wantRemaining: 15 * time.Minute,
		},
		{
			name:          "completed",
			duration:      time.Hour,
			progress:      ffmpeg.Progress{Speed: 2, Converted: time.Hour},
			wantCompleted: 1,
			wantRemaining: 0,
		},
		{
			name:          "zero speed",
			duration:      time.Hour,
			progress:      ffmpeg.Progress{Converted: 30 * time.Minute},
			prevSpeed:     2,
			wantCompleted: .5,
			wantRemaining: 15 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p := Progress{Duration: tt.duration}
			p.progress.Speed = tt.prevSpeed
			p.Update(tt.progress)
			assert.Equal(t, tt.wantCompleted, p.Completed())
			assert.Equal(t, tt.wantRemaining, p.Remaining())
		})
	}
}
