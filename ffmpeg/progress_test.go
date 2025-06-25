package ffmpeg

import (
	"log/slog"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFFMPEG_ProgressSocket(t *testing.T) {
	done := make(chan struct{})
	listener, sock, err := makeProgressSocket()
	require.NoError(t, err)

	handler := func(p Progress) {
		t.Helper()
		assert.Equal(t, time.Second, p.Converted)
		assert.InEpsilon(t, 1.0, p.Speed, 0.001)
		done <- struct{}{}
	}
	go serveProgressSocket(t.Context(), listener, sock, handler, slog.New(slog.DiscardHandler))

	fd, err := net.Dial("unix", sock)
	require.NoError(t, err)
	_, err = fd.Write([]byte("out_time_ms=1000000\nspeed=1.0x\nprogress=end\n"))
	require.NoError(t, err)
	<-done
}

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
			name:  "partial",
			input: "speed=1.1x\n",
			want:  []Progress{},
		},
		{
			name:  "partial",
			input: "out_time_ms=1000\n",
			want:  []Progress{},
		},
		{
			name:  "valid",
			input: "speed=1.1x\nout_time_ms=1000\n",
			want:  []Progress{{Converted: 1000000, Speed: 1.1}},
		},
		{
			name:  "multiple",
			input: "speed=1.0x\nout_time_ms=1\nspeed=1.1x\nout_time_ms=1000\n",
			want:  []Progress{{Converted: 1000, Speed: 1.0}, {Converted: 1000000, Speed: 1.1}},
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
	// Current:
	// Benchmark_progress-10    	     571	   2112925 ns/op	    4398 B/op	       7 allocs/op
	var input strings.Builder
	for range 1000 {
		for range 100 {
			input.WriteString("token=value\n")
		}
		input.WriteString("speed=1.1x\nout_time_ms=1\n")
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
