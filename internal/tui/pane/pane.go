package pane

import (
	tea "github.com/charmbracelet/bubbletea"
)

type Name string
type ActivateMsg struct {
	Pane Name
}

var _ tea.Model = Manager{}

type Manager struct {
	panes      map[Name]tea.Model
	activePane Name
}

func New(models map[Name]tea.Model) Manager {
	return Manager{panes: models}
}

func (m Manager) Init() tea.Cmd {
	initCommands := make([]tea.Cmd, 0, len(m.panes))
	for _, p := range m.panes {
		initCommands = append(initCommands, p.Init())
	}
	return tea.Batch(initCommands...)
}

func (m Manager) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case ActivateMsg:
		// set the active pane
		if _, ok := m.panes[msg.Pane]; ok {
			m.activePane = msg.Pane
		}
	case tea.KeyMsg:
		// keyMsg is only sent to the active pane
		if p, ok := m.panes[m.activePane]; ok {
			m.panes[m.activePane], cmd = p.Update(msg)
		}
	default:
		// broadcast message to all panes
		cmdList := make([]tea.Cmd, 0, len(m.panes))
		for paneName := range m.panes {
			var subCmd tea.Cmd
			m.panes[paneName], subCmd = m.panes[paneName].Update(msg)
			cmdList = append(cmdList, subCmd)
		}
		cmd = tea.Batch(cmdList...)
	}
	return m, cmd
}

func (m Manager) View() string {
	if p, ok := m.panes[m.activePane]; ok {
		return p.View()
	}
	return ""
}
