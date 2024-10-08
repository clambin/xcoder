package ffmpeg

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"iter"
	"math/rand"
	"net"
	"os"
	"os/exec"
	"path"
	"strconv"
	"time"
)

func (p Processor) Convert(ctx context.Context, request Request) error {
	if err := request.IsValid(); err != nil {
		return err
	}

	var sock string
	var err error
	if request.ProgressCB != nil {
		sock, err = p.progressSocket(request.ProgressCB)
		if err != nil {
			return fmt.Errorf("progress socket: %w", err)
		}
	}

	command, args, err := makeConvertCommand(request, sock)
	if err != nil {
		return err
	}

	cmd := exec.CommandContext(ctx, command, args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err = cmd.Run(); err != nil {
		err = fmt.Errorf("ffmpeg: %w. error: %s", err, lastLine(&stderr))
	}
	return err
}

// progressSocket creates and serves a unix socket for ffmpeg progress information.  Callers can use this to keep
// track of the progress of the conversion.
func (p Processor) progressSocket(progressCallback func(Progress)) (string, error) {
	// TODO: not sufficiently random?
	sockFileName := path.Join(os.TempDir(), "ffmpeg_socket_"+strconv.Itoa(rand.Int()))
	l, err := net.Listen("unix", sockFileName)
	if err != nil {
		return "", fmt.Errorf("progress socket: listen: %w", err)
	}
	go func() {
		defer func() {
			if err := os.Remove(sockFileName); err != nil {
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

func makeConvertCommand(request Request, progressSocket string) (string, []string, error) {
	videoCodec, ok := videoCodecs[request.VideoCodec]
	if !ok {
		return "", nil, fmt.Errorf("ffmpeg: unsupported video codec: %s", request.VideoCodec)
	}

	profile := "main"
	if request.BitsPerSample == 10 {
		profile = "main10"
	}

	args := []string{
		"-y",
		"-nostats", "-loglevel", "error",
	}
	args = append(args, prefixArguments...)
	args = append(args,
		//"-threads", "8",
		"-i", request.Source,
		"-map", "0",
		"-c:v", videoCodec, "-profile:v", profile, "-b:v", strconv.Itoa(request.BitRate),
		"-c:a", "copy",
		"-c:s", "copy",
		"-f", "matroska",
	)
	if progressSocket != "" {
		args = append(args, "-progress", "unix://"+progressSocket)
	}
	args = append(args, request.Target)
	return "ffmpeg", args, nil
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
			if bytes.Equal(line, endMarker) {
				return
			}
			if bytes.HasPrefix(line, convertedMarker) {
				if microSeconds, err := strconv.Atoi(string(line[len(convertedMarker):])); err == nil {
					prog.Converted = time.Duration(microSeconds) * time.Microsecond
					haveProgress = true
				}
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
		}
		if err := s.Err(); err != nil {
			yield(Progress{}, err)
		}
	}
}
