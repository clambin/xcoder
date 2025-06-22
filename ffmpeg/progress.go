package ffmpeg

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"time"
)

// makeProgressSocket creates and serves a unix socket for ffmpeg progress information.  Callers can use this to keep
// track of the progress of the conversion.
func makeProgressSocket() (net.Listener, string, error) {
	tmpDir, err := os.MkdirTemp("", "ffmpeg-")
	if err != nil {
		return nil, "", err
	}
	sockFileName := path.Join(tmpDir, "ffmpeg.sock")
	l, err := net.Listen("unix", sockFileName)
	if err != nil {
		return nil, "", fmt.Errorf("progress socket: listen: %w", err)
	}
	return l, sockFileName, nil
}

// serveProgressSocket reads the ffmpeg progress and calls the process callback function.
func serveProgressSocket(ctx context.Context, l net.Listener, path string, progressCallback func(Progress), logger *slog.Logger) {
	defer func() {
		if err := os.RemoveAll(filepath.Dir(path)); err != nil {
			logger.Error("failed to clean up status socket", "err", err)
		}
	}()

	fd, err := l.Accept()
	if err != nil {
		logger.Error("failed to serve status socket", "err", err)
		return
	}

	defer func() { _ = fd.Close() }()

	ch := progress(fd, logger)
	for {
		select {
		case prog, ok := <-ch:
			if !ok {
				logger.Debug("progress socket closed")
				return
			}
			progressCallback(prog)
		case <-ctx.Done():
			logger.Debug("context cancelled", "err", ctx.Err())
			return
		}
	}
}

type Progress struct {
	Converted time.Duration
	Speed     float64
}

// progress reads the ffmpeg progress information and returns Progress records on a channel
func progress(r io.Reader, logger *slog.Logger) chan Progress {
	var (
		convertedMarker = []byte("out_time_ms=")
		speedMarker     = []byte("speed=")
		endMarker       = []byte("progress=end")
	)
	ch := make(chan Progress)
	go func() {
		defer close(ch)
		s := bufio.NewScanner(r)
		var haveProgress, haveSpeed bool
		var prog Progress
		for s.Scan() {
			line := s.Bytes()
			switch {
			case bytes.HasPrefix(line, convertedMarker):
				microSeconds, _ := strconv.Atoi(string(line[len(convertedMarker):]))
				prog.Converted = time.Duration(microSeconds) * time.Microsecond
				haveProgress = true
			case bytes.HasPrefix(line, speedMarker):
				line = bytes.TrimSuffix(line, []byte("x"))
				prog.Speed, _ = strconv.ParseFloat(string(line[len(speedMarker):]), 64)
				haveSpeed = true
			}
			if haveProgress && haveSpeed {
				ch <- prog
				haveProgress = false
				haveSpeed = false
			}
			if bytes.Equal(line, endMarker) {
				return
			}
		}
		if err := s.Err(); err != nil {
			logger.Error("failed to read progress socket", "err", err)
		}
	}()
	return ch
}
