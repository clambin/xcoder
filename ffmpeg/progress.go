package ffmpeg

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"strconv"
	"strings"
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
			line := s.Text()
			args := strings.SplitN(line, "=", 2)
			if len(args) != 2 {
				logger.Warn("invalid progress line", "line", line)
				continue
			}
			switch args[0] {
			case "frame":
				prog.Frame, _ = strconv.ParseUint(args[1], 10, 64)
			case "fps":
				prog.FPS, _ = strconv.ParseFloat(args[1], 64)
			case "out_time_us":
				microSeconds, _ := strconv.Atoi(args[1])
				prog.Converted = time.Duration(microSeconds) * time.Microsecond
			case "dup_frames":
				prog.DuplicateFrames, _ = strconv.ParseUint(args[1], 10, 64)
			case "drop_frames":
				prog.DroppedFrames, _ = strconv.ParseUint(args[1], 10, 64)
			case "speed":
				prog.Speed, _ = strconv.ParseFloat(strings.TrimSuffix(args[1], "x"), 64)
			case "progress":
				ch <- prog
				if args[1] == "end" {
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
