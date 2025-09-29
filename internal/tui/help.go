package tui

import (
	"codeberg.org/clambin/bubbles/helper"
	"codeberg.org/clambin/bubbles/table"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// helpController is a helper that displays help for the active pane.
type helpController struct {
	panes      map[activePane]helper.Helper
	activePane activePane
}

func newHelpController(controllerKeyMap ControllerKeyMap, filterKeyMap help.KeyMap, styles helper.Styles) helpController {
	c := helpController{panes: make(map[activePane]helper.Helper)}

	// queue help
	h := controllerKeyMap.FullHelp()
	filterBindings := filterKeyMap.ShortHelp()
	filterBindings = append(filterBindings, table.DefaultFilterTableKeyMap().FilterKeyMap.ShortHelp()...)
	c.panes[queuePane] = helper.New().Styles(styles).Sections([]helper.Section{
		{Title: "General", Keys: h[0]},
		{Title: "View", Keys: h[1]},
		{Title: "Navigation", Keys: table.DefaultKeyMap().ShortHelp()},
		{Title: "Filters", Keys: filterBindings},
	})
	c.panes[logPane] = helper.New().Styles(styles).Sections([]helper.Section{
		{Title: "General", Keys: controllerKeyMap.ShortHelp()},
		{Title: "Navigation", Keys: []key.Binding{controllerKeyMap.CloseLogs}},
	})
	return c
}

func (c helpController) activeHelpPane() helper.Helper {
	return c.panes[c.activePane]
}

func (c helpController) Update(msg tea.Msg) (helpController, tea.Cmd) {
	switch msg := msg.(type) {
	case setPaneMsg:
		c.activePane = activePane(msg)
	}
	return c, nil
}
