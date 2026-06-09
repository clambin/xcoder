package ui

import (
	"charm.land/lipgloss/v2"
	"codeberg.org/clambin/bubbles/colors"
	"codeberg.org/clambin/bubbles/frame"
	"codeberg.org/clambin/bubbles/helper"
	"codeberg.org/clambin/bubbles/table"
)

var (
	defaultFrameStyle = frame.Style{
		Title:  lipgloss.NewStyle().Foreground(colors.Green),
		Border: lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(colors.Blue),
	}
)

type Styles struct {
	StatusStyles
	HelpStyles helper.Styles
	LogViewerStyles
	MediaViewerStyles
}

func DefaultStyles() Styles {
	return Styles{
		StatusStyles: StatusStyles{
			Main:       lipgloss.NewStyle().Foreground(colors.White).Background(colors.Blue),
			Processing: lipgloss.NewStyle().Foreground(lipgloss.Color("#8B0000")),
		},
		HelpStyles: helper.Styles{
			Header: lipgloss.NewStyle().Foreground(colors.Yellow).Italic(true),
			Key:    lipgloss.NewStyle().Foreground(colors.Yellow),
			Desc:   lipgloss.NewStyle().Foreground(colors.Grey58),
		},
		LogViewerStyles: LogViewerStyles{
			Frame: frame.Style{
				Title:  lipgloss.NewStyle().Foreground(colors.Green),
				Border: lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(colors.Blue),
			},
		},
		MediaViewerStyles: MediaViewerStyles{
			MediaViewerItemStyles: MediaViewerItemStyles{
				TableStyles: table.FilterTableStyles{
					Table: table.Styles{
						Header:   lipgloss.NewStyle().Foreground(colors.White).Bold(true),
						Selected: lipgloss.NewStyle().Foreground(colors.Blue).Background(colors.White),
						Cell:     lipgloss.NewStyle().Foreground(colors.Grey70),
					},
				},
				FrameStyle:       defaultFrameStyle,
				MediaFilterStyle: lipgloss.NewStyle().Foreground(colors.Magenta1),
				RowCountStyle:    lipgloss.NewStyle().Foreground(colors.White),
			},
			MediaViewerSessionsStyles: MediaViewerSessionsStyles{
				FrameStyle: defaultFrameStyle,
			},
		},
	}
}

type StatusStyles struct {
	Main       lipgloss.Style
	Processing lipgloss.Style
}

type LogViewerStyles struct {
	Frame frame.Style
}

type MediaViewerStyles struct {
	MediaViewerItemStyles
	MediaViewerSessionsStyles
}

type MediaViewerItemStyles struct {
	TableStyles      table.FilterTableStyles
	FrameStyle       frame.Style
	MediaFilterStyle lipgloss.Style
	RowCountStyle    lipgloss.Style
}

type MediaViewerSessionsStyles struct {
	FrameStyle frame.Style
}
