package ffmpeg

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math/rand"
	"net"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
)

// Processor implements video scanning (ffprobe) and converting (ffmpeg).  Implemented as a struct so that is can be
// mocked at the calling side.
type Processor struct {
	Logger *slog.Logger
}

func (p Processor) Probe(ctx context.Context, path string) (Probe, error) {
	output, err := p.runCommand(ctx,
		"ffprobe",
		"-v", "quiet",
		"-print_format", "json",
		"-show_format", "-show_streams",
		path,
	)
	if err != nil {
		return Probe{}, fmt.Errorf("probe: %w", err)
	}

	var probe Probe
	if err = json.NewDecoder(output).Decode(&probe); err != nil {
		err = fmt.Errorf("decode: %w", err)
	}

	return probe, err
}

func (p Processor) Convert(ctx context.Context, input, output, targetCodec string, bitrate int) error {
	return p.ConvertWithProgress(ctx, input, output, targetCodec, bitrate, nil)
}

func (p Processor) ConvertWithProgress(ctx context.Context, input, output, targetCodec string, bitrate int, progressCallback func(progress Progress)) error {
	var sock string
	var err error
	if progressCallback != nil {
		sock, err = p.createProgressSocket(progressCallback)
		if err != nil {
			return fmt.Errorf("progress socket: %w", err)
		}
	}

	command, args, err := makeConvertCommand(input, output, targetCodec, bitrate, sock)
	if err != nil {
		return err
	}

	stdout, err := p.runCommand(ctx, command, args...)
	if err != nil {
		err = fmt.Errorf("ffmpeg failed. output: %s. err: %w", stdout.String(), err)
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

		// w/out data trimming: 1885038999 ns/op
		// with data trimming:     3802577 ns/op
		if len(data) > 2*bufSize {
			data = data[bufSize:]
		}
	}
}

func (p Processor) runCommand(ctx context.Context, command string, args ...string) (*bytes.Buffer, error) {
	cmd := exec.CommandContext(ctx, command, args...)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	l := p.Logger.With(slog.String("cmd", command))

	l.Debug("running command")
	if err := cmd.Run(); err != nil {
		lines := strings.Split(stderr.String(), "\n")
		var lastLine string
		if len(lines) > 0 {
			lastLine = lines[len(lines)-1]
		}
		return nil, fmt.Errorf("command failed: %w. last line: %s", err, lastLine)
	}
	l.Debug("command successful")
	return &stdout, nil
}
