package ffmpeg

import (
	"bytes"
	"context"
	"github.com/stretchr/testify/assert"
	"log/slog"
	"testing"
)

func Test_runCommand(t *testing.T) {
	tests := []struct {
		name       string
		command    string
		args       []string
		wantErr    assert.ErrorAssertionFunc
		wantStdout string
		wantStderr string
	}{
		{
			name:       "stdout",
			command:    "bash",
			args:       []string{"-c", "echo foo"},
			wantErr:    assert.NoError,
			wantStdout: "foo\n",
		},
		{
			name:       "stderr",
			command:    "bash",
			args:       []string{"-c", "echo foo >&2"},
			wantErr:    assert.NoError,
			wantStderr: "foo\n",
		},
		{
			name:       "command failed",
			command:    "bash",
			args:       []string{"-c", "echo foo >&2; exit 1"},
			wantErr:    assert.Error,
			wantStderr: "foo\n",
		},
		{
			name:    "exec failed",
			command: "invalid",
			wantErr: assert.Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := Processor{Logger: slog.Default()}
			stdout, stderr, err := p.runCommand(context.Background(), tt.command, tt.args...)
			tt.wantErr(t, err)
			assert.Equal(t, tt.wantStdout, stdout.String())
			assert.Equal(t, tt.wantStderr, stderr.String())
		})
	}
}

func Test_lastLine(t *testing.T) {
	tests := []struct {
		name   string
		buffer *bytes.Buffer
		want   string
	}{
		{
			name: "one line",
			buffer: bytes.NewBufferString(`one
`),
			want: "one",
		},
		{
			name:   "one line (no linefeed)",
			buffer: bytes.NewBufferString(`one`),
			want:   "one",
		},
		{
			name: "multiline",
			buffer: bytes.NewBufferString(`one
two
three
four
`),
			want: "four",
		},
		{
			name: "multiline (no linefeed)",
			buffer: bytes.NewBufferString(`one
two
three
four`),
			want: "four",
		},
		{
			name:   "empty",
			buffer: &bytes.Buffer{},
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, lastLine(tt.buffer))
		})
	}
}
