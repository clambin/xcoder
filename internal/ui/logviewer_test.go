package ui

import (
	"github.com/stretchr/testify/assert"
	"log/slog"
	"testing"
)

func Test_logViewer(t *testing.T) {
	v := newLogViewer()

	l := slog.New(slog.NewTextHandler(v, &slog.HandlerOptions{}))
	l.Info("hello world")

	assert.Contains(t, v.GetText(true), "hello world")
}
