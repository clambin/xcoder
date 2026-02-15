package tui

import (
	"bytes"
	"io"
	"testing"
	"time"

	"github.com/charmbracelet/x/exp/teatest"
)

func waitFor(t testing.TB, r io.Reader, want []byte) {
	t.Helper()
	waitForFunc(t, r, func(b []byte) bool { return bytes.Contains(b, want) })
}

func waitForFunc(t testing.TB, r io.Reader, f func([]byte) bool) {
	t.Helper()
	teatest.WaitFor(t, r, f, teatest.WithDuration(time.Second), teatest.WithCheckInterval(10*time.Millisecond))
}
