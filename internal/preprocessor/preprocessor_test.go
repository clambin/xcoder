package preprocessor

import (
	"context"
	"errors"
	"github.com/clambin/videoConvertor/internal/ffmpeg"
	"github.com/clambin/videoConvertor/internal/profile"
	"github.com/clambin/videoConvertor/internal/worklist"
	"github.com/stretchr/testify/assert"
	"log/slog"
	"testing"
	"time"
)

func TestRun(t *testing.T) {
	type args struct {
		stats ffmpeg.VideoStats
		err   error
	}
	tests := []struct {
		name string
		args
		profile string
		want    any
	}{
		{
			name:    "video does not meet criteria",
			profile: "hevc-high",
			args: args{
				stats: ffmpeg.VideoStats{VideoCodec: "h264", Height: 720, BitRate: 1024, Duration: time.Hour, BitsPerSample: 10},
			},
			want: worklist.Rejected,
		},
		{
			name:    "video probe failed",
			profile: "hevc-high",
			args: args{
				err: errors.New("failed"),
			},
			want: worklist.Failed,
		},
		{
			name:    "video meets criteria",
			profile: "hevc-low",
			args: args{
				stats: ffmpeg.VideoStats{VideoCodec: "h264", Height: 720, BitRate: 1_024_000_000, Duration: time.Hour, BitsPerSample: 10},
			},
			want: worklist.Inspected,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ff := fakeFFMPEG{
				stats: tt.args.stats,
				err:   tt.args.err,
			}
			p, _ := profile.GetProfile(tt.profile)

			ctx, cancel := context.WithCancel(context.Background())
			done := make(chan struct{})
			ch := make(chan *worklist.WorkItem)
			go func() { Run(ctx, ch, &ff, p, slog.Default()); done <- struct{}{} }()

			item := worklist.WorkItem{Source: "foo.mkv"}
			ch <- &item
			assert.Eventually(t, func() bool {
				status, _ := item.Status()
				return status == tt.want
			}, time.Second, 10*time.Millisecond)

			cancel()
			<-done
		})
	}
}

var _ FFMPEG = &fakeFFMPEG{}

type fakeFFMPEG struct {
	stats ffmpeg.VideoStats
	err   error
}

func (f fakeFFMPEG) Scan(_ context.Context, _ string) (ffmpeg.VideoStats, error) {
	return f.stats, f.err
}
