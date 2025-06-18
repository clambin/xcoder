package convertor

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"iter"
	"log/slog"
	"net"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"time"

	"github.com/clambin/videoConvertor/ffmpeg"
)

// makeConvertCommand creates a exec.Command to run ffmeg with the required configuration.
func makeConvertCommand(ctx context.Context, request Request, progressSocket string) (*exec.Cmd, error) {
	codecName, ok := videoCodecs[request.TargetStats.VideoCodec]
	if !ok {
		return nil, fmt.Errorf("unsupported video codec: %s", request.TargetStats.VideoCodec)
	}
	profile := "main"
	if request.TargetStats.BitsPerSample == 10 {
		profile = "main10"
	}

	cmd := ffmpeg.Input(request.Source, inputArguments).
		Output(request.Target, ffmpeg.Args{
			"c:v":       codecName,
			"profile:v": profile,
			"b:v":       strconv.Itoa(request.TargetStats.BitRate),
			"c:a":       "copy",
			"c:s":       "copy",
			"f":         "matroska",
		}).
		NoStats().
		LogLevel("error").
		OverWriteTarget().
		ProgressSocket(progressSocket)

	return cmd.Build(ctx), nil
}

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
func serveProgressSocket(l net.Listener, path string, progressCallback func(Progress), logger *slog.Logger) {
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

	for prog, err := range progress(fd) {
		if err == nil {
			progressCallback(prog)
		} else {
			logger.Error("failed to process status socket", "err", err)
		}
	}
	_ = fd.Close()
}

type Progress struct {
	Converted time.Duration
	Speed     float64
}

// progress reads the ffmpeg progress information to create a complete Progress record and yield it to the caller.
func progress(r io.Reader) iter.Seq2[Progress, error] {
	var (
		convertedMarker = []byte("out_time_ms=")
		speedMarker     = []byte("speed=")
		endMarker       = []byte("progress=end")
	)
	return func(yield func(Progress, error) bool) {
		s := bufio.NewScanner(r)
		var haveProgress, haveSpeed bool
		var prog Progress
		for s.Scan() {
			line := s.Bytes()
			if bytes.HasPrefix(line, convertedMarker) {
				microSeconds, _ := strconv.Atoi(string(line[len(convertedMarker):]))
				prog.Converted = time.Duration(microSeconds) * time.Microsecond
				haveProgress = true
			} else if bytes.HasPrefix(line, speedMarker) {
				line = bytes.TrimSuffix(line, []byte("x"))
				prog.Speed, _ = strconv.ParseFloat(string(line[len(speedMarker):]), 64)
				haveSpeed = true
			}
			if haveProgress && haveSpeed {
				if !yield(prog, nil) {
					return
				}
				haveProgress = false
				haveSpeed = false
			}
			if bytes.Equal(line, endMarker) {
				return
			}
		}
		if err := s.Err(); err != nil {
			yield(Progress{}, err)
		}
	}
}
