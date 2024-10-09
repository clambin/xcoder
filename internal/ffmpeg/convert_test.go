package ffmpeg

import (
	"bytes"
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"log/slog"
	"net"
	"strings"
	"testing"
	"time"
)

func Test_makeConvertCommand(t *testing.T) {
	for _, tt := range makeConvertCommandTests {
		t.Run(tt.name, func(t *testing.T) {
			// ffmpeg-go.Silent() uses a global variable. :-(
			//t.Parallel()

			p := Processor{Logger: slog.Default()}

			type ctxKey string
			key := ctxKey("test")
			ctx := context.WithValue(context.Background(), key, "test")

			s, err := p.makeConvertCommand(ctx, tt.request, tt.progressSocket)
			tt.wantErr(t, err)
			if err != nil {
				return
			}

			clArgs := strings.Join(s.Compile().Args[1:], " ")
			assert.Equal(t, tt.want, clArgs)
			// check that the cmd will be run with our context
			assert.Equal(t, "test", s.Context.Value(key))
		})
	}
}

func TestProcessor_progressSocket(t *testing.T) {
	var p Processor
	done := make(chan struct{})
	sock, err := p.progressSocket(func(p Progress) {
		t.Helper()
		assert.Equal(t, time.Second, p.Converted)
		assert.Equal(t, 1.0, p.Speed)
		done <- struct{}{}
	})
	require.NoError(t, err)

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
			for p, err := range progress(bytes.NewReader([]byte(tt.input))) {
				if err == nil {
					progresses = append(progresses, p)
				}
			}
			assert.Equal(t, tt.want, progresses)
		})
	}
}

// Current:
// Benchmark_progress-16                607           1965400 ns/op         1233087 B/op          4 allocs/op
func Benchmark_progress(b *testing.B) {
	var input strings.Builder
	for range 1000 {
		for range 100 {
			input.WriteString("token=value\n")
		}
		input.WriteString("speed=1.1x\nout_time_ms=1\n")
	}
	input.WriteString("progress=end\n")
	buf := input.String()
	b.ResetTimer()
	for range b.N {
		for p, err := range progress(bytes.NewBufferString(buf)) {
			if err != nil {
				b.Fatal(err)
			}
			if p.Speed != 1.1 {
				b.Fatal("invalid speed")
			}
		}
	}
}
