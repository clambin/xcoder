package ffmpeg

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io"
	"log/slog"
	"testing"
)

func TestProcessor_readTempSocket(t *testing.T) {
	p := Processor{Logger: slog.Default()}

	var buffer bytes.Buffer
	assert.NoError(t, write(&buffer, 10, 5))

	var called int
	p.readProgressSocket(&buffer, func(progress Progress) {
		called++
	})
	assert.NotZero(t, called)
}

func Benchmark_serveSocket(b *testing.B) {
	var body bytes.Buffer
	assert.NoError(b, write(&body, 500, 100))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := Processor{Logger: slog.Default()}

		var called int
		p.readProgressSocket(bytes.NewBuffer(body.Bytes()), func(progress Progress) {
			called++
		})
		assert.NotZero(b, called)
	}
}

func write(conn io.Writer, batches, lines int) error {
	for range batches {
		for l := 0; l < lines; l++ {
			if _, err := conn.Write([]byte("other_info\n")); err != nil {
				return fmt.Errorf("write: %w", err)
			}
		}
		if _, err := conn.Write([]byte("speed=1.1x\n")); err != nil {
			return fmt.Errorf("write: %w", err)
		}
		if _, err := conn.Write([]byte("out_time_ms=1234\n")); err != nil {
			return fmt.Errorf("write: %w", err)
		}
	}
	_, err := conn.Write([]byte("progress=end\n"))
	if err != nil {
		err = fmt.Errorf("write: %w", err)
	}
	return err
}
