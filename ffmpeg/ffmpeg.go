// Package ffmpeg is a thin wrapper calling ffmpeg
package ffmpeg

import (
	"context"
	"os/exec"
	"slices"
)

type FFMPEG struct {
	inputArgs      Args
	outputArgs     Args
	globalPostArgs Args
	input          string
	output         string
}

func Input(path string, args Args) *FFMPEG {
	return &FFMPEG{
		input:          path,
		inputArgs:      args,
		outputArgs:     make(Args),
		globalPostArgs: make(Args),
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

func (f *FFMPEG) ProgressSocket(path string) *FFMPEG {
	if path != "" {
		f.globalPostArgs["progress"] = "unix://" + path
	} else {
		delete(f.globalPostArgs, "progress")
	}
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
	keys := make([]string, 0, len(a))
	for k := range a {
		keys = append(keys, k)
	}
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
