package ui

import (
	"io"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"codeberg.org/clambin/bubbles/helper"
	"codeberg.org/clambin/bubbles/table"
	"github.com/clambin/xcoder/internal/transcoder"
)

type WorkItems interface {
	Items() []*transcoder.WorkItem
}

var _ WorkItems = (*transcoder.WorkItems)(nil)

type Transcoder interface {
	Active() bool
	SessionCount() int
	SetActive(active bool)
	Subscribe() <-chan transcoder.SessionEvent
}

var _ Transcoder = (*transcoder.Transcoder)(nil)

var _ tea.Model = Application{}

type Application struct {
	keyMap      RootKeyMap
	helpWindow  helper.Helper
	logViewer   logViewer
	statusLine  statusLine
	mediaViewer mediaViewer
	// TODO
	height int
	windows
	mediaTableFilterIsOn bool
}

func New(workItems WorkItems, transcoder Transcoder, r io.Reader, keyMap KeyMap, styles Styles) Application {
	ch := transcoder.Subscribe()

	a := Application{
		statusLine:  newStatusLine(transcoder, styles.StatusStyles, spinner.WithSpinner(spinner.Meter)),
		logViewer:   newLogViewer(r, keyMap.LogViewerKeyMap, styles.LogViewerStyles),
		mediaViewer: newMediaViewer(workItems, transcoder, ch, keyMap.MediaViewerKeyMap, styles.MediaViewerStyles),
		keyMap:      keyMap.RootKeyMap,
	}
	var sections []helper.Section
	sections = append(sections, a.helpSections()...)

	a.helpWindow = helper.New().Sections(sections).Styles(styles.HelpStyles)

	return a
}

func (a Application) Init() tea.Cmd {
	return tea.Batch(
		a.mediaViewer.Init(),
		a.logViewer.Init(),
		a.statusLine.Init(),
	)
}

func (a Application) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return a.resize(msg.Width, msg.Height), nil
	case table.FilterStateChangeMsg:
		a.mediaTableFilterIsOn = msg.State
		// TODO
		a.mediaViewer, _ = a.mediaViewer.Update(msg)
		return a, nil
	case tea.KeyPressMsg:
		switch {
		// enable / disable these depending on the mediaWindow filterStatus
		case key.Matches(msg.Key(), a.keyMap.Quit) && !a.mediaTableFilterIsOn:
			return a, tea.Quit
		case key.Matches(msg.Key(), a.keyMap.Logs) && !a.mediaTableFilterIsOn:
			a.showLogs = !a.showLogs
			return a, nil
		case key.Matches(msg.Key(), a.keyMap.Help) && !a.mediaTableFilterIsOn:
			a.showHelp = !a.showHelp
			return a, nil
		default:
			var cmd tea.Cmd
			switch a.activeWindow() {
			case mediaWindow:
				a.mediaViewer, cmd = a.mediaViewer.Update(msg)
				return a, cmd
			case logWindow:
				a.logViewer, cmd = a.logViewer.Update(msg)
				return a, cmd
			case helpWindow:
				a.helpWindow, cmd = a.helpWindow.Update(msg)
				return a, cmd
			}
		}
	default:
		cmds := make([]tea.Cmd, 4)
		a.mediaViewer, cmds[0] = a.mediaViewer.Update(msg)
		a.logViewer, cmds[1] = a.logViewer.Update(msg)
		a.helpWindow, cmds[2] = a.helpWindow.Update(msg)
		a.statusLine, cmds[3] = a.statusLine.Update(msg)
		return a, tea.Batch(cmds...)
	}
	return a, nil
}

func (a Application) View() tea.View {
	var w string
	switch a.activeWindow() {
	case logWindow:
		w = a.logViewer.View()
	case mediaWindow:
		w = a.mediaViewer.View()
	case helpWindow:
		w = lipgloss.NewStyle().Height(a.height - 2).Render(a.helpWindow.View())
	}

	v := tea.NewView(
		lipgloss.JoinVertical(lipgloss.Top,
			w,
			a.statusLine.View(),
			a.helpWindow.ShortHelpView(a.shortHelp()),
		),
	)
	v.AltScreen = true
	return v
}

func (a Application) resize(width, height int) Application {
	a.height = height
	a.statusLine = a.statusLine.setWidth(width)
	a.helpWindow.SetWidth(width)
	height -= 2 // space for status & help lines
	a.logViewer = a.logViewer.SetSize(width, max(0, height))
	a.mediaViewer = a.mediaViewer.SetSize(width, max(0, height))
	return a
}

func (a Application) shortHelp() []key.Binding {
	bindings := a.keyMap.ShortHelp()
	switch a.activeWindow() {
	case logWindow:
		bindings = append(bindings, a.logViewer.keyMap.ShortHelp()...)
	case mediaWindow:
		bindings = append(bindings, a.mediaViewer.workItemsViewer.keyMap.ShortHelp()...)
	case helpWindow:
	}
	return bindings
}

func (a Application) helpSections() []helper.Section {
	sections := []helper.Section{{Title: "GENERAL", Keys: a.keyMap.FullHelp()[0]}}
	sections = append(sections, a.mediaViewer.helpSections()...)
	sections = append(sections, a.logViewer.helpSections()...)
	return sections
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// activeWindow represents the currently active window in the UI
type activeWindow int

const (
	mediaWindow activeWindow = iota
	logWindow
	helpWindow
)

// windows keeps track of the state of the UI windows
type windows struct {
	showLogs bool
	showHelp bool
}

func (w windows) activeWindow() activeWindow {
	switch {
	case w.showHelp:
		return helpWindow
	case w.showLogs:
		return logWindow
	default:
		return mediaWindow
	}
}
