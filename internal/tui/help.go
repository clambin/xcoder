package tui

import (
	"codeberg.org/clambin/bubbles/helper"
	"codeberg.org/clambin/bubbles/table"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/clambin/xcoder/internal/tui/pane"
)

// helpController is a helper that displays help for the active pane.
type helpController struct {
	panes      map[pane.Name]helper.Helper
	activePane pane.Name
}

func newHelpController(keyMap KeyMap, styles helper.Styles) helpController {
	c := helpController{panes: make(map[pane.Name]helper.Helper)}

	// queue help
	h := keyMap.Controller.FullHelp()
	filterBindings := keyMap.Filter.ShortHelp()
	filterBindings = append(filterBindings, table.DefaultFilterTableKeyMap().FilterKeyMap.ShortHelp()...)
	c.panes[queuePane] = helper.New().Styles(styles).Sections([]helper.Section{
		{Title: "General", Keys: h[0]},
		{Title: "View", Keys: h[1]},
		{Title: "Navigation", Keys: table.DefaultKeyMap().ShortHelp()},
		{Title: "Filters", Keys: filterBindings},
	})
	c.panes[logPane] = helper.New().Styles(styles).Sections([]helper.Section{
		{Title: "General", Keys: keyMap.Controller.ShortHelp()},
		{Title: "Navigation", Keys: []key.Binding{keyMap.LogViewer.CloseLogs}},
	})
	return c
}

func (c helpController) activeHelpPane() helper.Helper {
	return c.panes[c.activePane]
}

func (c helpController) Update(msg tea.Msg) (helpController, tea.Cmd) {
	switch msg := msg.(type) {
	case pane.ActivateMsg:
		c.activePane = msg.Pane
	}
	return c, nil
}

// TODO: add View().  This becomes its own tea.Model.
