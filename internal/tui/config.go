package tui

import (
	"strings"

	"github.com/clambin/xcoder/internal/pipeline"
)

// configPane displays the current configuration.
// Since the configuration is currently static, we pre-render it.
type configPane struct {
	content string
}

const (
	sourceLabel    = "Source"
	profileLabel   = "Profile"
	overwriteLabel = "Overwrite"
	removeLabel    = "Remove"
)

var configLabels = []string{sourceLabel, profileLabel, overwriteLabel, removeLabel}
var boolToString = map[bool]string{true: "on", false: "off"}

func newConfigPane(cfg pipeline.Configuration, styles ConfigStyles) configPane {
	var labelWidth int
	for _, label := range configLabels {
		labelWidth = max(labelWidth, len(label))
	}

	var out strings.Builder
	for _, label := range configLabels {
		//if i > 0 {
		out.WriteString("\n")
		//}
		out.WriteString(styles.LabelStyle.Render(label + strings.Repeat(" ", labelWidth-len(label)) + ": "))
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
		out.WriteString(styles.TextStyle.Render(value))
	}
	return configPane{content: out.String()}
}

func (c configPane) View() string {
	return c.content
}
