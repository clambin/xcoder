package ffmpeg

import (
	"io"
	"strings"
)

func lastLine(r io.Reader) string {
	content, _ := io.ReadAll(r)
	lines := strings.Split(string(content), "\n")
	for count := len(lines) - 1; count >= 0; count-- {
		if lines[count] != "" {
			return lines[count]
		}
	}
	return ""
}
