package ffmpeg

import (
	"bytes"
	"cmp"
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/exec"
	"strings"
)

type FFMPEG struct {
	progress           func(Progress)
	output             string
	progressSocketPath string
	args               []string
}

func Decode(path string, args ...string) *FFMPEG {
	f := new(FFMPEG)
	return f.Decode(path, args...)
}

func (ff *FFMPEG) Decode(path string, args ...string) *FFMPEG {
	ff.args = append(ff.args, args...)
	ff.args = append(ff.args, "-i", path)
	return ff
}

func (ff *FFMPEG) Encode(args ...string) *FFMPEG {
	ff.args = append(ff.args, args...)
	return ff
}

func (ff *FFMPEG) Muxer(muxer string, args ...string) *FFMPEG {
	ff.args = append(ff.args, "-f", muxer)
	if len(args) > 0 {
		ff.args = append(ff.args, args...)
	}
	return ff
}

func (ff *FFMPEG) Output(path string) *FFMPEG {
	ff.output = path
	return ff
}

func (ff *FFMPEG) LogLevel(level string) *FFMPEG {
	ff.args = append(ff.args, "-loglevel", level)
	return ff
}

func (ff *FFMPEG) NoStats() *FFMPEG {
	ff.args = append(ff.args, "-nostats")
	return ff
}

func (ff *FFMPEG) OverWriteTarget() *FFMPEG {
	ff.args = append(ff.args, "-y")
	return ff
}

func (ff *FFMPEG) Progress(cb func(Progress), progressSocketPath string) *FFMPEG {
	ff.progress = cb
	ff.progressSocketPath = progressSocketPath
	return ff
}

func (ff *FFMPEG) Build(ctx context.Context) *exec.Cmd {
	args := ff.args
	if ff.progress != nil {
		args = append(args, "-progress", "unix://"+ff.progressSocketPath)
	}
	args = append(args, cmp.Or(ff.output, "-"))
	return exec.CommandContext(ctx, "ffmpeg", args...)
}

func (ff *FFMPEG) Run(ctx context.Context) error {
	if ff.progressSocketPath != "" {
		l, err := net.Listen("unix", ff.progressSocketPath)
		if err != nil {
			return fmt.Errorf("listen: %w", err)
		}
		// TODO: race condition?  cmd may run before serve is blocking on Accept?
		go serveProgressSocket(ctx, l, ff.progressSocketPath, ff.progress, slog.New(slog.NewTextHandler(os.Stderr, nil)))
	}
	cmd := ff.Build(ctx)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("ffmpeg: %v (%s)", err, formatError(&stderr))
	}
	return err
}

func formatError(buf *bytes.Buffer) string {
	return strings.TrimSuffix(buf.String(), "\n")
	/*lines := strings.Split(strings.TrimSuffix(buf.String(), "\n"), "\n")
	if len(lines) == 0 {
		return ""
	}
	return lines[len(lines)-1]*/
}
