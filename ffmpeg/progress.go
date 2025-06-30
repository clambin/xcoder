package ffmpeg

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"strconv"
	"time"
)

// serveProgressSocket reads the ffmpeg progress and calls the process callback function.
func serveProgressSocket(ctx context.Context, l net.Listener, progressCallback func(Progress), logger *slog.Logger) error {
	fd, err := l.Accept()
	if err != nil {
		return fmt.Errorf("failed to serve status socket: %w", err)
	}

	defer func() { _ = fd.Close() }()

	ch := progress(fd, logger)
	for {
		select {
		case prog, ok := <-ch:
			if !ok {
				logger.Debug("progress socket closed")
				return nil
			}
			progressCallback(prog)
		case <-ctx.Done():
			logger.Debug("context cancelled", "err", ctx.Err())
			return nil
		}
	}
}

type Progress struct {
	Frame           uint64
	FPS             float64
	Converted       time.Duration
	DuplicateFrames uint64
	DroppedFrames   uint64
	Speed           float64
	// TBD: stream_0_0_q=-0.0
	// TBD: bitrate=N/A
	// TBD: total_size=N/A
}

// progress reads the ffmpeg progress information and returns Progress records on a channel.
func progress(r io.Reader, logger *slog.Logger) chan Progress {
	ch := make(chan Progress)
	go func() {
		defer close(ch)
		s := bufio.NewScanner(r)
		var prog Progress
		for s.Scan() {
			line := s.Bytes()
			parseProgressLine(line, []byte("frame="), &prog.Frame, nil)
			parseProgressLine(line, []byte("fps="), &prog.FPS, nil)
			var usec uint64
			if parseProgressLine(s.Bytes(), []byte("out_time_us="), &usec, nil) {
				prog.Converted = time.Duration(usec) * time.Microsecond
			}
			parseProgressLine(line, []byte("dup_frames="), &prog.DuplicateFrames, nil)
			parseProgressLine(line, []byte("drop_frames="), &prog.DroppedFrames, nil)
			parseProgressLine(line, []byte("speed="), &prog.Speed, []byte("x"))
			var p []byte
			if parseProgressLine(line, []byte("progress="), &p, nil) {
				ch <- prog
				if bytes.Equal(line, []byte("end")) {
					return
				}
			}
		}
		if err := s.Err(); err != nil {
			logger.Error("failed to read progress socket", "err", err)
		}
	}()
	return ch
}

func parseProgressLine(line []byte, prefix []byte, value any, suffix []byte) bool {
	arg, ok := bytes.CutPrefix(line, prefix)
	if !ok {
		return false
	}
	if len(suffix) > 0 {
		arg = bytes.TrimSuffix(arg, suffix)
	}
	switch p := value.(type) {
	case *float64:
		*p, _ = strconv.ParseFloat(string(arg), 64)
	case *uint64:
		*p, _ = strconv.ParseUint(string(arg), 10, 64)
	case *int:
		*p, _ = strconv.Atoi(string(arg))
	case *[]byte:
		*p = arg
	default:
		return false
	}
	return true
}
