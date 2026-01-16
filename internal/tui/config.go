package tui

import (
	"strings"

	"github.com/clambin/xcoder/internal/pipeline"
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

	var out strings.Builder
	for _, label := range configLabels {
		//if i > 0 {
		out.WriteString("\n")
		//}
		out.WriteString(styles.LabelStyle.Render(label + strings.Repeat(" ", labelWidth-len(label)) + ": "))
		var value string
		switch label {
		case sourceLabel:
			value = cfg.Input // TODO: if too long, truncate with ellipsis
		case profileLabel:
			value = cfg.ProfileName
		case overwriteLabel:
			value = boolToString[cfg.Overwrite]
		case removeLabel:
			value = boolToString[cfg.Remove]
		}
		out.WriteString(styles.TextStyle.Render(value))
	}
	return configView{content: out.String()}
}

func (c configView) View() string {
	return c.content
}
