package ffmpeg

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	ffmpeg "github.com/u2takey/ffmpeg-go"
	"io"
	"iter"
	"net"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"time"
)

// makeProgressSocket creates and serves a unix socket for ffmpeg progress information.  Callers can use this to keep
// track of the progress of the conversion.
func (p Processor) makeProgressSocket() (net.Listener, string, error) {
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

func (p Processor) serveProgressSocket(l net.Listener, path string, progressCallback func(Progress)) {
	defer func() {
		if err := os.RemoveAll(filepath.Dir(path)); err != nil {
			p.Logger.Error("failed to clean up status socket", "err", err)
		}
	}()

	fd, err := l.Accept()
	if err != nil {
		p.Logger.Error("failed to serve status socket", "err", err)
		return
	}

	for prog, err := range progress(fd) {
		if err == nil {
			progressCallback(prog)
		} else {
			p.Logger.Error("failed to process status socket", "err", err)
		}
	}
	_ = fd.Close()
}

func makeConvertCommand(ctx context.Context, request Request, progressSocket string) (*ffmpeg.Stream, error) {
	codecName, ok := videoCodecs[request.TargetVideoCodec]
	if !ok {
		return nil, fmt.Errorf("unsupported video codec: %s", request.TargetVideoCodec)
	}
	profile := "main"
	if request.SourceStats.BitsPerSample == 10 {
		profile = "main10"
	}

	globalArgs := []string{
		"-nostats",
		"-loglevel", "error",
	}
	if progressSocket != "" {
		globalArgs = append(globalArgs, "-progress", "unix://"+progressSocket)
	}
	outputArguments := ffmpeg.KwArgs{
		"c:v":       codecName,
		"profile:v": profile,
		"crf":       strconv.Itoa(request.ConstantRateFactor),
		"c:a":       "copy",
		"c:s":       "copy",
		"f":         "matroska",
	}

	cmd := ffmpeg.Input(request.Source, inputArguments).Output(request.Target, outputArguments).GlobalArgs(globalArgs...)
	cmd.Context = ctx
	cmd.OverWriteOutput() //.Silent(true) uses global variable, so not thread-safe.  See init()
	return cmd, nil
}

func init() {
	// ffmpeg-go's Silent() uses a global variable, so isn't thread-safe. So instead, we set the global variable here.
	ffmpeg.LogCompiledCommand = false
}

type Progress struct {
	Converted time.Duration
	Speed     float64
}

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
