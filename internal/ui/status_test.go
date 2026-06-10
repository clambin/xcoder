package ui

import (
	"strings"
	"testing"
	"unicode/utf8"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStatusLine_BatchStatus(t *testing.T) {
	const expectedWidth = 92
	var x fakeTranscoder
	s := newStatusLine(&x, StatusStyles{}).setWidth(expectedWidth)
	for _, msg := range flattenBatchCmd(s.Init()()) {
		s, _ = s.Update(msg)
	}

	tests := []struct {
		status bool
		want   string
	}{
		{true, "Overwrite target: ON  Remove source: ON  Batch processing:"},
		{true, "Overwrite target: ON  Remove source: ON  Batch processing: ON"},
		{true, "Overwrite target: ON  Remove source: ON  Batch processing:"},
		{false, "Overwrite target: ON  Remove source: ON  Batch processing: OFF"},
		{false, "Overwrite target: ON  Remove source: ON  Batch processing: OFF"},
	}

	for idx, tt := range tests {
		x.active = tt.status
		s, _ = s.Update(blinkStatusMsg{})
		got := s.View()
		require.Len(t, got, expectedWidth)
		assert.Equal(t, tt.want, strings.TrimSpace(got), idx)

	}
}

func TestStatusLine_Converting(t *testing.T) {
	const expectedWidth = 95
	transcoder := fakeTranscoder{
		active: true,
		count:  2,
	}

	s := newStatusLine(&transcoder, StatusStyles{}, spinner.WithSpinner(spinner.Dot)).setWidth(expectedWidth)

	v := s.View()
	assert.Equal(t, expectedWidth, utf8.RuneCountInString(ansi.Strip(v)))
	assert.Equal(t, "  Converting 2 file(s) ... ⣾   Overwrite target: ON  Remove source: ON  Batch processing:      ", v)
	s, _ = s.Update(s.spinner.Tick())
	s, _ = s.Update(blinkStatusMsg{})
	v = s.View()
	assert.Equal(t, expectedWidth, utf8.RuneCountInString(ansi.Strip(v)))
	assert.Equal(t, "  Converting 2 file(s) ... ⣽   Overwrite target: ON  Remove source: ON  Batch processing: ON   ", v)
}

func flattenBatchCmd(msg tea.Msg) []tea.Msg {
	if cmd, ok := msg.(tea.BatchMsg); ok {
		msgs := make([]tea.Msg, len(cmd))
		for i, m := range cmd {
			msgs[i] = m()
		}
		return msgs
	}
	return []tea.Msg{msg}
}
