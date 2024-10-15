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
	assert.Zero(t, l.Size())
	assert.False(t, l.Active())
	for _ = range l.All() {
		t.Errorf("should not be reached")
	}

	l.Add("foo")
	assert.Equal(t, 1, l.Size())
	var count int
	for _ = range l.All() {
		count++
	}
	assert.Equal(t, l.Size(), count)

}

func TestWorkList_NextToConvert(t *testing.T) {
	var l WorkList

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
	stats := ffmpeg.VideoStats{VideoCodec: "h264", Height: 1080, BitRate: 5_000_000}
	var item WorkItem
	item.AddSourceStats(stats)
	assert.Equal(t, stats, item.SourceVideoStats())
}

func TestWorkItem_RemainingFormatted(t *testing.T) {
	tests := []struct {
		name     string
		status   WorkStatus
		input    time.Duration
		expected string
	}{
		{"No duration", Converting, 0, ""},
		{"Only seconds", Converting, 10 * time.Second, "10s"},
		{"Only minutes", Converting, 5 * time.Minute, "5m"},
		{"Only hours", Converting, 2 * time.Hour, "2h"},
		{"Days and hours", Converting, 26 * time.Hour, "1d2h"},
		{"Hours and minutes", Converting, 3*time.Hour + 45*time.Minute, "3h45m"},
		{"Hours, minutes, seconds", Converting, 1*time.Hour + 30*time.Minute + 20*time.Second, "1h30m20s"},
		{"Hours and seconds", Converting, 1*time.Hour + 2*time.Second, "1h2s"},
		{"Exactly two days", Converting, 48 * time.Hour, "2d"},
		{"Multiple units", Converting, 72*time.Hour + 10*time.Minute + 15*time.Second, "3d10m15s"},
		{"Not converting", Converted, 72*time.Hour + 10*time.Minute + 15*time.Second, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var item WorkItem
			item.SetStatus(tt.status, nil)
			item.Progress.Duration = 2 * tt.input
			item.Progress.Update(ffmpeg.Progress{Converted: tt.input, Speed: 1})
			assert.Equal(t, tt.expected, item.RemainingFormatted())
		})
	}
}

func TestWorkItem_CompletedFormatted(t *testing.T) {
	tests := []struct {
		name   string
		status WorkStatus
		input  time.Duration
		want   string
	}{
		{"not converting", Converted, time.Second, ""},
		{"starting", Converting, 0, ""},
		{"half done", Converting, 30 * time.Minute, "50.0%"},
		{"done", Converting, time.Hour, "100.0%"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var item WorkItem
			item.SetStatus(tt.status, nil)
			item.Progress.Duration = time.Hour
			item.Progress.Update(ffmpeg.Progress{Converted: tt.input, Speed: 1})
			assert.Equal(t, tt.want, item.CompletedFormatted())
		})
	}
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
