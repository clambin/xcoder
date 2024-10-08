package ffmpeg

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
)

func (p Processor) Convert(ctx context.Context, request Request) error {
	if err := request.IsValid(); err != nil {
		return err
	}

	var sock string
	var err error
	if request.ProgressCB != nil {
		sock, err = p.createProgressSocket(request.ProgressCB)
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

func (p Processor) createProgressSocket(progressCallback func(Progress)) (string, error) {
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

		p.readProgressSocket(fd, progressCallback)
		_ = fd.Close()
	}()
	return sockFileName, nil
}

func (p Processor) readProgressSocket(conn io.Reader, progressCallback func(Progress)) {
	var data string

	const bufSize = 256
	buf := make([]byte, bufSize)
	for !strings.Contains(data, "progress=end") {
		n, err := conn.Read(buf)
		if err != nil {
			if !errors.Is(err, io.EOF) {
				p.Logger.Error("failed to read from progress socket", "err", err)
			}
			return
		}
		data += string(buf[:n])

		if progress, ok := getProgress(data); ok {
			progressCallback(progress)
		}

		// oversize buffer:       2904326 ns/op         4793430 B/op       6614 allocs/op
		// smarter:               1364488 ns/op         2001133 B/op       6411 allocs/op

		// last line may have been incomplete. process it after we get more data.
		if pos := strings.LastIndexByte(data, '\n'); pos != -1 {
			data = data[pos+1:]
		}
	}
}

func makeConvertCommand(request Request, progressSocket string) (string, []string, error) {
	ffmpegArgs, err := getConvertArgsByOS(request)
	if err != nil {
		return "", nil, err
	}

	if progressSocket != "" {
		ffmpegArgs = append(ffmpegArgs,
			"-progress", "unix://"+progressSocket,
		)
	}

	ffmpegArgs = append(ffmpegArgs, request.Target)
	return "ffmpeg", ffmpegArgs, nil
}
