package ui

import (
	"fmt"
	"github.com/clambin/videoConvertor/internal/configuration"
	"github.com/clambin/videoConvertor/internal/worklist"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"strings"
)

type header struct {
	*configPane
	*statusPane
	*shortcutsView
	*tview.Grid
}

func newHeader(list *worklist.WorkList, configuration configuration.Configuration) *header {
	h := header{
		Grid:          tview.NewGrid(),
		configPane:    newConfigPane(configuration),
		statusPane:    newStatusPane(list),
		shortcutsView: newShortcutsView(),
	}
	h.Grid.AddItem(h.configPane.TextView, 0, 0, 1, 1, 0, 0, false)
	h.Grid.AddItem(h.statusPane.TextView, 0, 1, 1, 1, 0, 0, false)
	h.Grid.AddItem(h.shortcutsView.Pages, 0, 2, 1, 2, 0, 0, false)
	return &h
}

func (h *header) refresh() {
	h.statusPane.refresh()
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type configPane struct {
	*tview.TextView
}

func newConfigPane(cfg configuration.Configuration) *configPane {
	p := configPane{TextView: tview.NewTextView().SetDynamicColors(true)}
	var content strings.Builder
	content.WriteString(fmt.Sprintf(" [%s]Base directory:   [%s]%s\n", labelColor, tview.Styles.SecondaryTextColor, cfg.Input))
	content.WriteString(fmt.Sprintf(" [%s]Profile:          [%s]%s\n", labelColor, tview.Styles.SecondaryTextColor, cfg.ProfileName))
	content.WriteString(fmt.Sprintf(" [%s]Remove source:    [%s]%s\n", labelColor, tview.Styles.SecondaryTextColor, onOffString[cfg.RemoveSource]))
	content.WriteString(fmt.Sprintf(" [%s]Overwrite target: [%s]%s\n", labelColor, tview.Styles.SecondaryTextColor, onOffString[cfg.OverwriteNewerTarget]))
	p.TextView.SetText(content.String())
	return &p
}

var onOffString = map[bool]string{
	true:  "on",
	false: "off",
}

var onOffColor = map[bool]tcell.Color{
	true:  tcell.ColorGreen,
	false: textColor,
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type statusPane struct {
	*tview.TextView
	list *worklist.WorkList
}

func newStatusPane(list *worklist.WorkList) *statusPane {
	return &statusPane{
		TextView: tview.NewTextView().SetDynamicColors(true),
		list:     list,
	}
}

func (s *statusPane) refresh() {
	converting := s.list.Active()
	content := fmt.Sprintf(" [%s]Converting    : [%s]%s\n", labelColor, onOffColor[converting], onOffString[converting])
	s.TextView.Clear()
	s.TextView.SetText(content)
}
