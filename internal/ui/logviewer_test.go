package ui

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/charmbracelet/x/exp/golden"
)

func TestLogViewer(t *testing.T) {
	const lineCount = 3
	var r bytes.Buffer
	for i := range lineCount {
		r.WriteString(fmt.Sprintf("line %d\n", i+1))
	}
	v := newLogViewer(&r, LogViewerKeyMap{}, LogViewerStyles{}).SetSize(40, 10)

	v, cmd := v.Update(v.Init()())
	for range lineCount {
		v, cmd = v.Update(cmd())
	}

	golden.RequireEqual(t, v.View())
}
