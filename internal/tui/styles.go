package tui

import (
	"codeberg.org/clambin/bubbles/colors"
	"codeberg.org/clambin/bubbles/frame"
	"codeberg.org/clambin/bubbles/helper"
	"codeberg.org/clambin/bubbles/table"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/lipgloss"
	"github.com/clambin/xcoder/internal/pipeline"
)

var (
	statusColors = map[string]lipgloss.Style{
		pipeline.Rejected.String():   lipgloss.NewStyle().Foreground(colors.IndianRed),
		pipeline.Skipped.String():    lipgloss.NewStyle().Foreground(colors.Green),
		pipeline.Converted.String():  lipgloss.NewStyle().Foreground(colors.Yellow4Alt),
		pipeline.Converting.String(): lipgloss.NewStyle().Foreground(colors.Orange1),
		pipeline.Failed.String():     lipgloss.NewStyle().Foreground(colors.Red),
	}
)

type Styles struct {
	Config      ConfigStyles
	Help        helper.Styles
	TableStyle  table.FilterTableStyles
	FrameStyle  frame.Styles
	Status      StatusStyles
	MediaFilter lipgloss.Style
}

type ConfigStyles struct {
	LabelStyle lipgloss.Style
	TextStyle  lipgloss.Style
}

type StatusStyles struct {
	Main       lipgloss.Style
	Processing lipgloss.Style
}

func DefaultStyles() Styles {
	return Styles{
		FrameStyle: frame.Styles{
			Title:  lipgloss.NewStyle().Foreground(colors.Green),
			Border: lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(colors.Blue),
		},
		Config: ConfigStyles{
			LabelStyle: lipgloss.NewStyle().Foreground(colors.Orange1),
			TextStyle:  lipgloss.NewStyle().Foreground(colors.Grey50),
		},
		Help: helper.Styles{
			Header: lipgloss.NewStyle().Foreground(colors.Green),
			Key:    lipgloss.NewStyle().Foreground(colors.IndianRed),
			Desc:   lipgloss.NewStyle().Foreground(colors.Grey50),
		},
		Status: StatusStyles{
			Main:       lipgloss.NewStyle().Background(colors.Blue),
			Processing: lipgloss.NewStyle().Foreground(lipgloss.Color("#8B0000")),
		},
		TableStyle: table.FilterTableStyles{
			Table: table.Styles{
				Header:   lipgloss.NewStyle().Foreground(colors.White),
				Selected: lipgloss.NewStyle().Foreground(colors.White).Background(colors.Blue),
				Cell:     lipgloss.NewStyle().Foreground(colors.Blue),
				Frame: frame.Styles{
					Title:  lipgloss.NewStyle().Foreground(colors.Green),
					Border: lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(colors.Blue),
				},
			},
			Filter: table.FilterStyles{
				Border: lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(colors.Blue),
				TextArea: textarea.Style{
					Base:             lipgloss.Style{},
					CursorLine:       lipgloss.Style{},
					CursorLineNumber: lipgloss.Style{},
					EndOfBuffer:      lipgloss.Style{},
					LineNumber:       lipgloss.Style{},
					Placeholder:      lipgloss.Style{},
					Prompt:           lipgloss.Style{},
					Text:             lipgloss.Style{},
				},
			},
		},
		MediaFilter: lipgloss.NewStyle().Foreground(colors.DeepPink1Alt).Bold(true).Italic(true),
	}
}
