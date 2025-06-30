package ffmpeg

import (
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_progress(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []Progress
	}{
		{
			name:  "empty",
			input: "",
			want:  []Progress{},
		},
		{
			name:  "invalid",
			input: "foo\nprogress=end\n",
			want:  []Progress{{}},
		},
		{
			name:  "partial",
			input: "speed=1.1x\n",
			want:  []Progress{},
		},
		{
			name:  "partial",
			input: "speed=1.1x\nout_time_us=1000\n",
			want:  []Progress{},
		},
		{
			name:  "valid",
			input: "speed=1.1x\nout_time_us=1000\nprogress=end\n",
			want:  []Progress{{Converted: time.Millisecond, Speed: 1.1}},
		},
		{
			name:  "multiple",
			input: "speed=1.0x\nout_time_us=1\nprogress=continue\nspeed=1.1x\nout_time_us=2\nprogress=end\n",
			want:  []Progress{{Converted: time.Microsecond, Speed: 1.0}, {Converted: 2 * time.Microsecond, Speed: 1.1}},
		},
		{
			name:  "full",
			input: "frame=10\nfps=25.0\nout_time_us=1000\ndup_frames=1\ndrop_frames=2\nspeed=10x\nprogress=end\n",
			want:  []Progress{{Frame: 10, FPS: 25.0, Converted: time.Millisecond, DuplicateFrames: 1, DroppedFrames: 2, Speed: 10}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			progresses := make([]Progress, 0, len(tt.want))
			for p := range progress(strings.NewReader(tt.input), slog.New(slog.DiscardHandler)) {
				progresses = append(progresses, p)
			}
			assert.Equal(t, tt.want, progresses)
		})
	}
}

func Benchmark_progress(b *testing.B) {
	// Benchmark_progress-10    	    5250	    215401 ns/op	    4290 B/op	       4 allocs/op
	var input strings.Builder
	for range 1000 {
		input.WriteString("frame=10\nfps=25.0\nout_time_us=1000\ndup_frames=1\ndrop_frames=2\nspeed=1.1x\n")
	}
	input.WriteString("progress=end\n")
	buf := input.String()
	l := slog.New(slog.DiscardHandler)
	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		for p := range progress(strings.NewReader(buf), l) {
			if p.Speed != 1.1 {
				b.Fatal("invalid speed")
			}
		}
	}
}
