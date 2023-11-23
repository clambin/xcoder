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
	for b := 0; b < batches; b++ {
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

/*
func TestBufferedRead(t *testing.T) {
	const oneMB = 1024 * 1024
	const batches = 10
	s := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Length", strconv.Itoa(batches*oneMB))
		for i := 0; i < batches; i++ {
			time.Sleep(250 * time.Millisecond)
			_, _ = writer.Write([]byte(strings.Repeat("x", oneMB)))
		}
	}))
	defer s.Close()

	req, _ := http.NewRequest(http.MethodGet, s.URL, nil)
	t.Log("calling")
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func(Body io.ReadCloser) { _ = Body.Close() }(resp.Body)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	t.Log("request completed")

	block := make([]byte, oneMB)
	var rcvd int
	for {
		n, err := resp.Body.Read(block)
		if n > 0 {
			rcvd += n
			t.Log("block received: " + strconv.Itoa(n))
		}
		if errors.Is(err, io.EOF) {
			break
		}
	}
	assert.Equal(t, batches*oneMB, rcvd)
}
*/
