package cmd

import (
	"context"
	"os/exec"
	"slices"
	"strings"
)

type FFMPEG struct {
	source         string
	sourceArgs     Args
	target         string
	targetArgs     Args
	globalPostArgs Args
}

func Input(path string, args Args) *FFMPEG {
	f := &FFMPEG{}
	return f.Input(path, args)
}

func (f *FFMPEG) maybeInit() {
	if f.sourceArgs == nil {
		f.sourceArgs = make(Args)
	}
	if f.targetArgs == nil {
		f.targetArgs = make(Args)
	}
	if f.globalPostArgs == nil {
		f.globalPostArgs = make(Args)
	}
}

func (f *FFMPEG) Input(path string, args Args) *FFMPEG {
	f.source = path
	f.sourceArgs = args
	return f
}

func (f *FFMPEG) Output(path string, args Args) *FFMPEG {
	f.target = path
	f.targetArgs = args
	return f
}

func (f *FFMPEG) LogLevel(level string) *FFMPEG {
	f.maybeInit()
	f.globalPostArgs["loglevel"] = level
	return f
}

func (f *FFMPEG) NoStats() *FFMPEG {
	f.maybeInit()
	f.globalPostArgs["nostats"] = ""
	return f
}

func (f *FFMPEG) OverWriteTarget(overwrite bool) *FFMPEG {
	f.maybeInit()
	if overwrite {
		f.globalPostArgs["y"] = ""
	} else {

	}
	return f
}

func (f *FFMPEG) ProgressSocket(path string) *FFMPEG {
	f.maybeInit()
	if path != "" {
		f.globalPostArgs["progress"] = "unix://" + path
	} else {
		delete(f.globalPostArgs, "progress")
	}
	return f
}

func (f *FFMPEG) AddGlobalArguments(args Args) *FFMPEG {
	f.maybeInit()
	for k, v := range args {
		f.globalPostArgs[k] = v
	}
	return f
}

func (f *FFMPEG) Build(ctx context.Context) *exec.Cmd {
	return exec.CommandContext(ctx, "ffmeg",
		f.sourceArgs.compile(),
		"-i", f.source,
		f.targetArgs.compile(),
		f.target,
		f.globalPostArgs.compile(),
	)
}

type Args map[string]string

func (a Args) compile() string {
	arguments := make([]string, 0, len(a))
	for k, v := range a {
		if v != "" {
			arguments = append(arguments, "-"+k+" "+v)
		} else {
			arguments = append(arguments, "-"+k)
		}
	}
	slices.Sort(arguments)
	return strings.Join(arguments, " ")
}
