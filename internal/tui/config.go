package tui

import (
	"unicode/utf8"

	"github.com/charmbracelet/lipgloss"
	"github.com/clambin/xcoder/internal/pipeline"
)

const (
	maxValueWidth = 30
)

// configView displays the current configuration.
// Since the configuration is currently static, we pre-render it.
type configView struct {
	content string
}

const (
	sourceLabel    = "Source"
	profileLabel   = "Profile"
	overwriteLabel = "Overwrite"
	removeLabel    = "Remove"
)

var configLabels = []string{sourceLabel, profileLabel, overwriteLabel, removeLabel}
var boolToString = map[bool]string{true: "active", false: "off"}

func newConfigView(cfg pipeline.Configuration, styles ConfigStyles) configView {
	var labelWidth int
	for _, label := range configLabels {
		labelWidth = max(labelWidth, len(label))
	}

	parts := make([]string, len(configLabels))
	for i, label := range configLabels {
		var value string
		switch label {
		case sourceLabel:
			value = cfg.Input
		case profileLabel:
			value = cfg.ProfileName
		case overwriteLabel:
			value = boolToString[cfg.Overwrite]
		case removeLabel:
			value = boolToString[cfg.Remove]
		}
		parts[i] = lipgloss.JoinHorizontal(lipgloss.Left,
			styles.LabelStyle.Width(labelWidth+2).Render(label+": "),
			styles.TextStyle.Render(truncateLeft(value, maxValueWidth)),
		)
	}
	return configView{content: lipgloss.NewStyle().
		Padding(1, 0, 0, 0).
		Render(lipgloss.JoinVertical(lipgloss.Top, parts...)),
	}
}

func (c configView) View() string {
	return c.content
}

func truncateLeft(s string, maxWidth int) string {
	if lipgloss.Width(s) <= maxWidth {
		return s
	}

	const ellipsis = "…"
	ellW := lipgloss.Width(ellipsis)

	if ellW >= maxWidth {
		return ellipsis
	}

	// Strip runes from the left until it fits
	for lipgloss.Width(s)+ellW > maxWidth {
		_, size := utf8.DecodeRuneInString(s)
		s = s[size:]
	}

	return ellipsis + s
}
