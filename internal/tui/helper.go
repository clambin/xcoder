package tui

import (
	"codeberg.org/clambin/bubbles/helper"
)

// helpViewer displays the help given an active
type helpViewer struct {
	content map[paneID]helper.Helper
	styles  helper.Styles
}

func newHelpViewer(content map[paneID]helper.Helper, styles helper.Styles) helpViewer {
	return helpViewer{content: content, styles: styles}
}

func (h *helpViewer) view(activePane paneID, width int) string {
	return h.content[activePane].Styles(h.styles).Width(width).View()
}
