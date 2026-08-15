package ui

import (
	"strings"
	"testing"
	"unicode/utf8"

	"charm.land/bubbles/v2/spinner"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStatusBar_BatchStatus(t *testing.T) {
	const expectedWidth = 92
	var x fakeTranscoder
	s := newStatusBar(&x, "test", StatusStyles{}).Width(expectedWidth)

	tests := []struct {
		status bool
		want   string
	}{
		{true, "Profile: test Overwrite target: ON Remove source: ON Batch processing: ON   "},
		{true, "Profile: test Overwrite target: ON Remove source: ON Batch processing:      "},
		{true, "Profile: test Overwrite target: ON Remove source: ON Batch processing: ON   "},
		{false, "Profile: test Overwrite target: ON Remove source: ON Batch processing: OFF  "},
		{false, "Profile: test Overwrite target: ON Remove source: ON Batch processing: OFF  "},
	}

	for idx, tt := range tests {
		x.SetActive(tt.status)
		s, _ = s.Update(refreshStatusBarMsg{})
		got := s.View()
		require.Len(t, got, expectedWidth)
		assert.Equal(t, tt.want, strings.TrimLeft(got, " "), idx)
	}
}

func TestStatusBar_Converting(t *testing.T) {
	const expectedWidth = 108
	transcoder := fakeTranscoder{count: 2}
	transcoder.SetActive(true)

	styles := DefaultStyles().StatusStyles
	s := newStatusBar(&transcoder, "test", styles, spinner.WithSpinner(spinner.Dot)).Width(expectedWidth)

	v := ansi.Strip(s.View())
	assert.Equal(t, expectedWidth, utf8.RuneCountInString(v))
	assert.Equal(t, "Converting 2 file(s) ... ⣾    Profile: test Overwrite target: ON Remove source: ON Batch processing:      ", strings.TrimLeft(v, " "))

	s, _ = s.Update(s.statusbar.Init()())
	s, _ = s.Update(refreshStatusBarMsg{})

	v = ansi.Strip(s.View())
	assert.Equal(t, expectedWidth, utf8.RuneCountInString(v))
	assert.Equal(t, "Converting 2 file(s) ... ⣽    Profile: test Overwrite target: ON Remove source: ON Batch processing: ON   ", strings.TrimLeft(v, " "))
}

func BenchmarkStatusBar_refresh(b *testing.B) {
	// BenchmarkStatusBar_refresh-10    	  102093	     11683 ns/op	    3832 B/op	      86 allocs/op
	transcoder := fakeTranscoder{count: 2}
	transcoder.SetActive(true)

	styles := DefaultStyles().StatusStyles
	s := newStatusBar(&transcoder, "test", styles, spinner.WithSpinner(spinner.Dot)).Width(180)

	b.ReportAllocs()
	for b.Loop() {
		s = s.refresh()
		_ = s.View()
	}
}
