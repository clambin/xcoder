package ffmpeg

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"iter"
	"log/slog"
	"net"
	"strconv"
	"time"
)

// serveProgressSocket reads the ffmpeg progress and calls the process callback function.
func serveProgressSocket(l net.Listener, progressCallback func(Progress), logger *slog.Logger) error {
	fd, err := l.Accept()
	if err != nil {
		return fmt.Errorf("failed to serve status socket: %w", err)
	}

	defer func() { _ = fd.Close() }()

	for prog := range progress(fd, logger) {
		progressCallback(prog)
	}
	logger.Debug("progress socket closed")
	return nil
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

func progress(r io.Reader, logger *slog.Logger) iter.Seq[Progress] {
	return func(yield func(Progress) bool) {
		s := bufio.NewScanner(r)
		var prog Progress
		for s.Scan() {
			line := s.Bytes()
			key, val, ok := bytes.Cut(line, []byte("="))
			if !ok {
				continue
			}

			var err error
			switch string(key) {
			case "frame":
				prog.Frame, err = strconv.ParseUint(string(val), 10, 64)
			case "fps":
				prog.FPS, err = strconv.ParseFloat(string(val), 64)
			case "out_time_us":
				var usec uint64
				usec, err = strconv.ParseUint(string(val), 10, 64)
				if err == nil {
					prog.Converted = time.Duration(usec) * time.Microsecond
				}
			case "dup_frames":
				prog.DuplicateFrames, err = strconv.ParseUint(string(val), 10, 64)
			case "drop_frames":
				prog.DroppedFrames, err = strconv.ParseUint(string(val), 10, 64)
			case "speed":
				v := bytes.TrimSuffix(val, []byte("x"))
				prog.Speed, err = strconv.ParseFloat(string(v), 64)
			case "progress":
				if !yield(prog) {
					return
				}
				if bytes.Equal(val, []byte("end")) {
					return
				}
			}
			if err != nil {
				logger.Error("failed to parse progress line", "line", string(line), "err", err)
			}
		}
	}
}
