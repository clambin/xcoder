package ffmpeg

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	ffmpeg "github.com/u2takey/ffmpeg-go"
	"io"
	"iter"
	"net"
	"os"
	"path"
	"strconv"
	"time"
)

type Request struct {
	Source      string
	Target      string
	TargetStats VideoStats
	ProgressCB  func(Progress)
}

var ErrMissingFilename = errors.New("missing filename")
var ErrInvalidCodec = errors.New("only hevc supported")
var ErrInvalidBitsPerSample = errors.New("bits per sample must be 8 or 10")
var ErrInvalidBitRate = errors.New("invalid bitrate")

func (r Request) IsValid() error {
	if r.Source == "" || r.Target == "" {
		return ErrMissingFilename
	}
	if r.TargetStats.VideoCodec != "hevc" {
		return ErrInvalidCodec
	}
	if r.TargetStats.BitsPerSample != 8 && r.TargetStats.BitsPerSample != 10 {
		return ErrInvalidBitsPerSample
	}
	if r.TargetStats.BitRate == 0 {
		return ErrInvalidBitRate
	}
	return nil
}

func (p Processor) Convert(ctx context.Context, request Request) error {
	if err := request.IsValid(); err != nil {
		return err
	}

	var sock string
	if request.ProgressCB != nil {
		var err error
		if sock, err = p.progressSocket(request.ProgressCB); err != nil {
			return fmt.Errorf("progress socket: %w", err)
		}
	}
	stream, err := makeConvertCommand(ctx, request, sock)
	if err != nil {
		return fmt.Errorf("failed to create command: %w", err)
	}
	p.Logger.Info("converting", "cmd", stream.Compile().String())
	return stream.Run()
}

// progressSocket creates and serves a unix socket for ffmpeg progress information.  Callers can use this to keep
// track of the progress of the conversion.
func (p Processor) progressSocket(progressCallback func(Progress)) (string, error) {
	tmpDir, err := os.MkdirTemp("", "ffmpeg-")
	if err != nil {
		return "", err
	}
	sockFileName := path.Join(tmpDir, "ffmpeg.sock")
	l, err := net.Listen("unix", sockFileName)
	if err != nil {
		return "", fmt.Errorf("progress socket: listen: %w", err)
	}
	go func() {
		defer func() {
			if err := os.RemoveAll(tmpDir); err != nil {
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
	}()
	return sockFileName, nil
}

func makeConvertCommand(ctx context.Context, request Request, progressSocket string) (*ffmpeg.Stream, error) {
	codecName, ok := videoCodecs[request.TargetStats.VideoCodec]
	if !ok {
		return nil, fmt.Errorf("unsupported video codec: %s", request.TargetStats.VideoCodec)
	}
	profile := "main"
	if request.TargetStats.BitsPerSample == 10 {
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
		//"map":       "0:0",
		"c:v":       codecName,
		"profile:v": profile,
		"b:v":       request.TargetStats.BitRate,
		"c:a":       "copy",
		"c:s":       "copy",
		"f":         "matroska",
	}

	cmd := ffmpeg.Input(request.Source, inputArguments).Output(request.Target, outputArguments).GlobalArgs(globalArgs...)
	cmd.Context = ctx
	cmd.OverWriteOutput() //.Silent(true)
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
