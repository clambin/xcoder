// Package ffmpeg is a thin wrapper calling ffmpeg
package ffmpeg

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"maps"
	"os/exec"
	"slices"
)

type FFMPEG struct {
	inputArgs      Args
	outputArgs     Args
	globalPostArgs Args
	logger         *slog.Logger
	input          string
	output         string
}

func Input(path string, args Args) *FFMPEG {
	return &FFMPEG{
		input:          path,
		inputArgs:      args,
		outputArgs:     make(Args),
		globalPostArgs: make(Args),
		logger:         slog.Default(),
	}
}

func (f *FFMPEG) Output(path string, args Args) *FFMPEG {
	f.output = path
	f.outputArgs = args
	return f
}

func (f *FFMPEG) LogLevel(level string) *FFMPEG {
	f.globalPostArgs["loglevel"] = level
	return f
}

func (f *FFMPEG) NoStats() *FFMPEG {
	f.globalPostArgs["nostats"] = ""
	return f
}

func (f *FFMPEG) OverWriteTarget() *FFMPEG {
	f.globalPostArgs["y"] = ""
	return f
}

func (f *FFMPEG) WithLogger(logger *slog.Logger) *FFMPEG {
	f.logger = logger
	return f
}

func (f *FFMPEG) ProgressSocket(ctx context.Context, cb func(Progress)) *FFMPEG {
	listener, path, err := makeProgressSocket()
	if err != nil {
		panic("failed to make progress socket: " + err.Error())
	}
	go serveProgressSocket(ctx, listener, path, cb, f.logger)
	f.globalPostArgs["progress"] = "unix://" + path
	return f
}

func (f *FFMPEG) AddGlobalArguments(args Args) *FFMPEG {
	for k, v := range args {
		f.globalPostArgs[k] = v
	}
	return f
}

func (f *FFMPEG) Build(ctx context.Context) *exec.Cmd {
	args := f.inputArgs.compile()
	args = append(args, "-i", f.input)
	args = append(args, f.outputArgs.compile()...)
	args = append(args, f.output)
	args = append(args, f.globalPostArgs.compile()...)

	return exec.CommandContext(ctx, "ffmpeg", args...)
}

type Args map[string]string

func (a Args) compile() []string {
	keys := slices.Collect(maps.Keys(a))
	slices.Sort(keys)

	arguments := make([]string, 0, 2*len(a))
	for _, k := range keys {
		arguments = append(arguments, "-"+k)
		if v := a[k]; v != "" {
			arguments = append(arguments, v)
		}
	}
	return arguments
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func Probe(path string) (VideoStats, error) {
	cmd := exec.Command("ffprobe",
		"-show_format",
		"-show_streams",
		"-loglevel", "error",
		"-output_format", "json",
		path,
	)
	var stdOut, stdErr bytes.Buffer
	cmd.Stdout = &stdOut
	cmd.Stderr = &stdErr
	if err := cmd.Run(); err != nil {
		return VideoStats{}, fmt.Errorf("[%s] %w", stdErr.String(), err)
	}
	return parseVideoStats(&stdOut)
}
