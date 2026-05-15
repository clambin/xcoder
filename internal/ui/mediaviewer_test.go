package ui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ltrim(t *testing.T) {
	tests := []struct {
		name  string
		input string
		n     int
		want  string
	}{
		{"empty", "", 10, ""},
		{"short", "012", 10, "012"},
		{"exact", "0123456789", 10, "0123456789"},
		{"long", "012345678912", 10, "…345678912"},
		{"unicode short", "012🤷", 10, "012🤷"},
		{"unicode exact", "012345678🤷", 10, "012345678🤷"},
		{"unicode long", "012345678912🤷", 10, "…45678912🤷"},
		{"no space", "0123", 0, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ltrim(tt.input, tt.n, '…')
			assert.Equal(t, tt.want, got)
		})
	}
}
