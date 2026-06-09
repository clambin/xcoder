package transcoder

import (
	"context"
	"fmt"
	"log/slog"
	"sync/atomic"
	"testing"
	"time"

	"github.com/clambin/xcoder/ffmpeg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTranscoder_AddMediaFile(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	t.Cleanup(cancel)

	var q WorkItems
	var cfg Configuration
	cfg.Profile, _ = GetProfile("hevc-high")
	//l := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))
	l := slog.New(slog.DiscardHandler)
	transcoder := New(&q, cfg, l)
	transcoder.controller.(*engine).probeFunc = func(path string) (ffmpeg.VideoStats, error) {
		time.Sleep(1000 * time.Millisecond)
		sourceStats := ffmpeg.VideoStats{Height: 1080, BitRate: 6_000_000, VideoCodec: "h264"}
		switch path {
		case "file_0.mkv":
			return ffmpeg.VideoStats{}, assert.AnError
		case "file_1.mkv":
			sourceStats.VideoCodec = "hevc"
		case "file_2.mkv":
			sourceStats.BitRate = 2_000_000
		case "file_3.mkv":
			sourceStats.Height = 720
		}
		return sourceStats, nil
	}
	go func() { _ = transcoder.Run(ctx) }()

	const (
		fileCount        = 5
		successFileCount = 1
	)

	for i := range fileCount {
		transcoder.AddMediaFile(fmt.Sprintf("file_%d.mkv", i))
	}

	assert.Eventually(t, func() bool {
		return len(q.Items()) == fileCount && len(q.ItemsWithStatus(StatusScanned)) == successFileCount
	}, 5*time.Second, 10*time.Millisecond)
}

func BenchmarkTranscoder_AddMediaFile(b *testing.B) {
	ctx, cancel := context.WithCancel(b.Context())
	b.Cleanup(cancel)

	var q WorkItems
	var cfg Configuration
	cfg.Profile, _ = GetProfile("hevc-high")
	//l := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))
	l := slog.New(slog.DiscardHandler)
	transcoder := New(&q, cfg, l)
	transcoder.controller.(*engine).probeFunc = func(path string) (ffmpeg.VideoStats, error) {
		sourceStats := ffmpeg.VideoStats{Height: 1080, BitRate: 6_000_000, VideoCodec: "h264"}
		return sourceStats, nil
	}
	go func() { _ = transcoder.Run(ctx) }()

	var i int
	b.ReportAllocs()
	for b.Loop() {
		transcoder.AddMediaFile(fmt.Sprintf("file_%d.mkv", i))
		i++
	}
}

func TestTranscoder_Transcode(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	t.Cleanup(cancel)

	const fileCount = 5
	var q WorkItems
	for i := range fileCount {
		item := WorkItem{
			Source: File{Path: fmt.Sprintf("test_%d.mkv", i)},
			Target: File{Path: fmt.Sprintf("test_%d.hevc.mkv", i)},
		}
		item.SetStatus(StatusScanned, nil)
		q.Add(&item)
	}

	var cfg Configuration
	cfg.Profile, _ = GetProfile("hevc-high")
	//l := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))
	l := slog.New(slog.DiscardHandler)
	transcoder := New(&q, cfg, l)
	transcoder.controller.(*engine).probeFunc = func(path string) (ffmpeg.VideoStats, error) {
		sourceStats := ffmpeg.VideoStats{Height: 1080, BitRate: 3_000_000, VideoCodec: "hevc"}
		return sourceStats, nil
	}
	transcoder.controller.(*engine).transcodeFunc = func(session *Session) error {
		time.Sleep(scheduleInterval * 2)
		return nil
	}
	transcoder.SetActive(true)

	var started, stopped atomic.Int64
	ch := transcoder.Subscribe()
	go func() {
		for ev := range ch {
			switch ev.Type {
			case SessionStartedEvent:
				started.Add(1)
			case SessionStoppedEvent:
				stopped.Add(1)
			}
		}
	}()

	go func() { _ = transcoder.Run(ctx) }()

	// wait for all sessions to complete
	require.Eventually(t, func() bool {
		return started.Load() == fileCount && stopped.Load() == fileCount
	}, 5*time.Second, 10*time.Millisecond)

	// all source files should be done
	assert.Len(t, q.ItemsWithStatus(StatusDone), fileCount)

	// all target files should be added to the worklist, scanned and marked as skipped
	// (as they're in the target codec).
	assert.Len(t, q.ItemsWithStatus(StatusSkipped), fileCount)
}

func Test_processSessionProgress(t *testing.T) {
	tests := []struct {
		name          string
		duration      time.Duration
		progress      ffmpeg.Progress
		expectedSpeed float64
		expectedETA   time.Duration
	}{
		{
			name:          "start",
			duration:      time.Hour,
			progress:      ffmpeg.Progress{Converted: 0, Speed: 2},
			expectedSpeed: 2,
			expectedETA:   30 * time.Minute,
		},
		{
			name:          "midway",
			duration:      time.Hour,
			progress:      ffmpeg.Progress{Converted: 30 * time.Minute, Speed: 2},
			expectedSpeed: 2,
			expectedETA:   15 * time.Minute,
		},
		{
			name:          "finished",
			duration:      time.Hour,
			progress:      ffmpeg.Progress{Converted: time.Hour, Speed: 2},
			expectedSpeed: 2,
			expectedETA:   0,
		},
		{
			name:          "stalled",
			duration:      time.Hour,
			progress:      ffmpeg.Progress{Converted: 0, Speed: 0},
			expectedSpeed: 0,
			expectedETA:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := Session{WorkItem: &WorkItem{Source: File{VideoStats: ffmpeg.VideoStats{Duration: tt.duration}}}}
			speed, eta := processSessionProgress(&session, tt.progress)
			assert.Equal(t, tt.expectedSpeed, speed)
			assert.Equal(t, tt.expectedETA, eta)
		})
	}
}
