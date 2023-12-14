package ffmpeg

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"testing"
)

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
