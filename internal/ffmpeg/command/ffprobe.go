package command

import (
	"bytes"
	"fmt"
	"os/exec"
)

func Probe(path string) (string, error) {
	args := []string{
		"-show_format",
		"-show_streams",
		"-loglevel", "error",
		"-output_format", "json",
		path,
	}

	cmd := exec.Command("ffprobe", args...)
	var stdOut, stdErr bytes.Buffer
	cmd.Stdout = &stdOut
	cmd.Stderr = &stdErr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("[%s] %w", stdErr.String(), err)
	}
	return stdOut.String(), nil
}
