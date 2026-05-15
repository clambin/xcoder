package ui

import (
	"fmt"
	"maps"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/progress"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"codeberg.org/clambin/bubbles/frame"
	"codeberg.org/clambin/bubbles/helper"
	"codeberg.org/clambin/bubbles/table"
	"codeberg.org/clambin/bubbles/ticker"
	"github.com/clambin/xcoder/internal/transcoder"
)

const (
	tableRefreshInterval  = 250 * time.Millisecond
	tickerRefreshInterval = time.Second
)

// mediaViewer is the main window of the UI.  It displays the media files and running transcode sessions
type mediaViewer struct {
	transcodeSessionsViewer transcodeSessionsViewer
	workItemsViewer         workItemsViewer
}

func newMediaViewer(
	workItems WorkItems,
	transcoder Transcoder,
	ch <-chan transcoder.SessionEvent,
	keyMap MediaViewerKeyMap,
	styles MediaViewerStyles,
) mediaViewer {
	return mediaViewer{
		workItemsViewer: workItemsViewer{
			FilterTable: table.NewFilterTable().
				Columns(workItemsColumns).
				KeyMap(table.DefaultFilterTableKeyMap()).
				Styles(styles.TableStyles),
			workItems:  workItems,
			transcoder: transcoder,
			styles:     styles.MediaViewerItemStyles,
			keyMap:     keyMap,
		},
		transcodeSessionsViewer: transcodeSessionsViewer{
			ch:     ch,
			styles: styles.MediaViewerSessionsStyles,
		},
	}
}

func (v mediaViewer) Init() tea.Cmd {
	return tea.Batch(
		v.workItemsViewer.Init(),
		v.transcodeSessionsViewer.Init(),
	)
}

func (v mediaViewer) Update(msg tea.Msg) (mediaViewer, tea.Cmd) {
	cmds := make([]tea.Cmd, 2)
	v.workItemsViewer, cmds[0] = v.workItemsViewer.Update(msg)
	v.transcodeSessionsViewer, cmds[1] = v.transcodeSessionsViewer.Update(msg)
	return v, tea.Batch(cmds...)
}

func (v mediaViewer) View() string {
	panes := make([]string, 1, 2)
	panes[0] = v.workItemsViewer.View()
	if len(v.transcodeSessionsViewer.sessions) > 0 {
		panes = append(panes, v.transcodeSessionsViewer.View())
	}
	return lipgloss.JoinVertical(lipgloss.Top, panes...)
}

func (v mediaViewer) SetSize(width, height int) mediaViewer {
	v.workItemsViewer = v.workItemsViewer.SetSize(width, height)
	v.transcodeSessionsViewer = v.transcodeSessionsViewer.Width(width)
	return v
}

func (v mediaViewer) helpSections() []helper.Section {
	return []helper.Section{
		{Title: "MEDIA", Keys: v.workItemsViewer.keyMap.FullHelp()[0]},
		{Title: "MEDIA FILTER", Keys: v.workItemsViewer.keyMap.FullHelp()[1]},
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

var workItemsColumns = []table.Column{
	{Name: "Source"},
	{Name: "Source Stats", Width: 20},
	//{Name: "Target"},
	{Name: "Target Stats", Width: 20},
	{Name: "Status", Width: 12},
	{Name: "Error"},
}

// workItemsViewer displays the list of media files and their status.
// Items can be filtered by status and can be manually queued for transcoding
type workItemsViewer struct {
	styles     MediaViewerItemStyles
	workItems  WorkItems
	transcoder Transcoder
	keyMap     MediaViewerKeyMap
	table.FilterTable
	sessionCount         int
	width                int
	height               int
	mediaFilterState     mediaFilterState
	showFullPath         bool
	mediaTableFilterIsOn bool
}

func (v workItemsViewer) Init() tea.Cmd {
	return tea.Batch(
		v.FilterTable.Init(),
		tea.Tick(tableRefreshInterval, func(t time.Time) tea.Msg {
			return refreshTable{}
		}),
	)
}

func (v workItemsViewer) Update(msg tea.Msg) (workItemsViewer, tea.Cmd) {
	switch msg := msg.(type) {
	case refreshTable:
		return v, tea.Batch(
			refreshTableCmd(v.workItems.Items(), v.mediaFilterState, v.showFullPath),
			tea.Tick(tableRefreshInterval, func(t time.Time) tea.Msg {
				return refreshTable{}
			}),
		)
	case setRowsMsg:
		v.FilterTable = v.Rows(msg.rows)
		return v, nil
	case table.FilterStateChangeMsg:
		v.mediaTableFilterIsOn = msg.State
		return v, refreshTableCmd(v.workItems.Items(), v.mediaFilterState, v.showFullPath)
	case transcodeSessionEventMsg:
		// we keep our own count as we can't be sure transcodeSessionEventMsg reaches transcodeSessionsViewer before we do.
		switch msg.Type {
		case transcoder.SessionStartedEvent:
			v.sessionCount++
		case transcoder.SessionStoppedEvent:
			v.sessionCount--
		}
		return v.SetSize(v.width, v.height), nil
	case tea.KeyPressMsg:
		// if the text filter is on, we ignore all key presses and route directly to the table.
		if !v.mediaTableFilterIsOn {
			switch {
			case key.Matches(msg, v.keyMap.ShowFullPath):
				v.showFullPath = !v.showFullPath
				return v, refreshTableCmd(v.workItems.Items(), v.mediaFilterState, v.showFullPath)
			case key.Matches(msg, v.keyMap.HideSkippedFiles):
				v.mediaFilterState.hideSkipped = !v.mediaFilterState.hideSkipped
				return v, refreshTableCmd(v.workItems.Items(), v.mediaFilterState, v.showFullPath)
			case key.Matches(msg, v.keyMap.HideRejectedFiles):
				v.mediaFilterState.hideRejected = !v.mediaFilterState.hideRejected
				return v, refreshTableCmd(v.workItems.Items(), v.mediaFilterState, v.showFullPath)
			case key.Matches(msg, v.keyMap.AutoProcess):
				v.transcoder.SetActive(!v.transcoder.Active())
				return v, nil
			case key.Matches(msg, v.keyMap.ConvertSelected):
				if row := v.SelectedRow(); row != nil {
					userData := row[len(row)-1].(table.UserData)
					item, ok := userData.Data.(*transcoder.WorkItem)
					if !ok {
						panic("selected row is not a work item")
					}
					status, _ := item.Status()
					if status != transcoder.StatusScanned {
						return v, nil
					}
					item.SetStatus(transcoder.StatusQueued, nil)
				}
				return v, nil
			}
		}
		var cmd tea.Cmd
		v.FilterTable, cmd = v.FilterTable.Update(msg)
		return v, cmd

	default:
		var cmd tea.Cmd
		v.FilterTable, cmd = v.FilterTable.Update(msg)
		return v, cmd
	}
}

func (v workItemsViewer) View() string {
	return frame.Render("media files", lipgloss.Center, v.styles.FrameStyle, v.FilterTable.View())
}

func (v workItemsViewer) SetSize(width, height int) workItemsViewer {
	v.width, v.height = width, height
	borderWidth, borderHeight := v.styles.FrameStyle.BorderSize()
	width -= borderWidth
	height -= borderHeight
	if v.sessionCount > 0 {
		// note: this assumes workItemsViewer and transcodeSessionsViewer both have borders or both have no borders
		height -= v.sessionCount + borderHeight
	}
	v.FilterTable = v.Size(max(0, width), max(0, height))
	return v
}

// loadTableCmd builds the table with the current Queue state and issues a command to load it in the table.
func refreshTableCmd(items []*transcoder.WorkItem, f mediaFilterState, showFullPath bool) tea.Cmd {
	return func() tea.Msg {
		rows := make([]table.Row, 0, len(items))
		for _, item := range items {
			if f.Show(item) {
				rows = append(rows, itemToRow(item, showFullPath))
			}
		}
		return setRowsMsg{rows: rows}
	}
}

func itemToRow(item *transcoder.WorkItem, showFullPath bool) table.Row {
	source := item.Source.Path
	if !showFullPath {
		source = filepath.Base(source)
	}
	status, err := item.Status()
	var errString string
	if err != nil {
		errString = err.Error()
	}

	return table.Row{
		source,
		item.Source.VideoStats.String(),
		item.Target.VideoStats.String(),
		status.String(),
		errString,
		table.UserData{Data: item},
	}
}

// mediaFilterState holds the current state of the mediaFilter. It determines which media files should be shown/hidden.
type mediaFilterState struct {
	hideSkipped   bool
	hideRejected  bool
	hideConverted bool
}

// Show returns true if the given item should be shown
func (s mediaFilterState) Show(item *transcoder.WorkItem) bool {
	status, _ := item.Status()
	switch status {
	case transcoder.StatusRejected:
		return !s.hideRejected
	case transcoder.StatusSkipped:
		return !s.hideSkipped
	case transcoder.StatusDone:
		return !s.hideConverted
	default:
		return true
	}
}

// String returns a string representation of the mediaFilterState
func (s mediaFilterState) String() string {
	on := map[string]struct{}{"skipped": {}, "rejected": {}, "converted": {}}
	if s.hideSkipped {
		delete(on, "skipped")
	}
	if s.hideRejected {
		delete(on, "rejected")
	}
	if s.hideConverted {
		delete(on, "converted")
	}
	if len(on) == 3 {
		return ""
	}
	if len(on) == 0 {
		return "none"
	}
	onString := slices.Collect(maps.Keys(on))
	slices.Sort(onString)
	return strings.Join(onString, ",")
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type transcodeSessionsViewer struct {
	styles   MediaViewerSessionsStyles
	ch       <-chan transcoder.SessionEvent
	sessions []transcodeSession
	width    int
}

func (v transcodeSessionsViewer) Init() tea.Cmd {
	return tea.Batch(
		tea.Tick(tickerRefreshInterval, func(_ time.Time) tea.Msg {
			return refreshTranscodeSessionsMsg{}
		}),
		func() tea.Msg {
			return transcodeSessionEventMsg(<-v.ch)
		},
	)
}

func (v transcodeSessionsViewer) Update(msg tea.Msg) (transcodeSessionsViewer, tea.Cmd) {
	switch msg := msg.(type) {
	case refreshTranscodeSessionsMsg:
		for i := range v.sessions {
			v.sessions[i].update()
		}
		return v, tea.Tick(tickerRefreshInterval, func(t time.Time) tea.Msg {
			return refreshTranscodeSessionsMsg{}
		})
	case transcodeSessionEventMsg:
		switch msg.Type {
		case transcoder.SessionStartedEvent:
			v.sessions = append(v.sessions, transcodeSession{
				session:  msg.Session,
				speed:    ticker.Ticker{MaxValues: 255},
				progress: progress.New(progress.WithDefaultBlend()),
				width:    v.width,
			})
		case transcoder.SessionStoppedEvent:
			v.sessions = slices.DeleteFunc(v.sessions, func(s transcodeSession) bool {
				return s.session == msg.Session
			})
		}
		return v, func() tea.Msg { return transcodeSessionEventMsg(<-v.ch) }
	default:
		return v, nil
	}
}

func (v transcodeSessionsViewer) View() string {
	sessions := make([]string, len(v.sessions))
	for i, session := range v.sessions {
		sessions[i] = session.view()
	}
	content := lipgloss.JoinVertical(lipgloss.Top, sessions...)
	return frame.Render("transcoder sessions", lipgloss.Center, v.styles.FrameStyle, content)
}

func (v transcodeSessionsViewer) Width(width int) transcodeSessionsViewer {
	borderWidth, _ := v.styles.FrameStyle.BorderSize()
	v.width = max(0, width-borderWidth)
	for i := range v.sessions {
		v.sessions[i].width = v.width
	}
	return v
}

type transcodeSession struct {
	progress progress.Model
	session  *transcoder.Session
	speed    ticker.Ticker
	width    int
}

func (s *transcodeSession) update() {
	s.speed = s.speed.Add(s.session.Progress().Speed)
}

func (s *transcodeSession) view() string {
	const speedWidth = 15
	const progressWidth = 20
	p := s.session.Progress()
	remainingWidth := s.width

	// speed
	speed := " Speed: " + s.speed.Width(speedWidth).View() + fmt.Sprintf(" %4.1fx", p.Speed)
	remainingWidth -= lipgloss.Width(speed)

	// progress
	s.progress.SetWidth(progressWidth)
	pctComplete := p.Converted.Seconds() / s.session.WorkItem.Source.VideoStats.Duration.Seconds()
	prog := " Progress: " + s.progress.ViewAs(pctComplete)
	remainingWidth -= lipgloss.Width(prog)

	// eta
	eta := max(0, (s.session.WorkItem.Source.VideoStats.Duration.Seconds()-p.Converted.Seconds())/p.Speed)
	etaLabel := fmt.Sprintf(" ETA: %-8s", (time.Duration(eta) * time.Second).String())
	remainingWidth -= lipgloss.Width(etaLabel)

	// now that we have all the fixed-size labels, we can trim the filename if necessary
	filename := filepath.Base(s.session.WorkItem.Source.Path)
	if len(filename) > remainingWidth {
		filename = ltrim(filename, max(0, remainingWidth), '…')
	}

	return lipgloss.JoinHorizontal(lipgloss.Left,
		lipgloss.NewStyle().Width(remainingWidth).Render(filename),
		speed,
		prog,
		etaLabel,
	)
}

func ltrim(s string, n int, trim rune) string {
	if n == 0 {
		return ""
	}
	r := []rune(s)
	rl := len(r)
	if rl <= n {
		return s
	}
	return string(append([]rune{trim}, r[rl-n+1:]...))
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// messages generated by mediaViewe and children

// refreshTable tells workItemsViewer to refresh the table
type refreshTable struct{}

type setRowsMsg struct {
	rows []table.Row
}

// refreshTranscodeSessionsMsg tells transcodeSessionsViewer to refresh the transcode sessions
type refreshTranscodeSessionsMsg struct{}

// transcodeSessionEventMsg indicates that a transcoder session has been started or stopped.
type transcodeSessionEventMsg transcoder.SessionEvent
