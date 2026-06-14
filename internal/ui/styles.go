package ui

import (
	"charm.land/lipgloss/v2"
	"codeberg.org/clambin/bubbles/colors"
	"codeberg.org/clambin/bubbles/frame"
	"codeberg.org/clambin/bubbles/helper"
	"codeberg.org/clambin/bubbles/table"
	"github.com/clambin/xcoder/internal/transcoder"
)

var (
	defaultFrameStyle = frame.Style{
		Title:  lipgloss.NewStyle().Foreground(colors.Green),
		Border: lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(colors.Blue),
	}

	statusStyles = map[string]lipgloss.Style{
		transcoder.StatusRejected.String():    lipgloss.NewStyle().Foreground(colors.IndianRed),
		transcoder.StatusSkipped.String():     lipgloss.NewStyle().Foreground(colors.Yellow4Alt),
		transcoder.StatusTranscoding.String(): lipgloss.NewStyle().Foreground(colors.Orange1),
		transcoder.StatusFailed.String():      lipgloss.NewStyle().Foreground(colors.Red),
		transcoder.StatusConverted.String():   lipgloss.NewStyle().Foreground(colors.Green4),
	}

	lightWhite = colors.Grey70
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
			Main:       lipgloss.NewStyle().Foreground(colors.Black).Background(colors.Blue),
			Processing: lipgloss.NewStyle().Foreground(colors.Red).Background(colors.Blue),
		},
		HelpStyles: helper.Styles{
			Header: lipgloss.NewStyle().Foreground(colors.Yellow).Italic(true),
			Key:    lipgloss.NewStyle().Foreground(colors.Yellow),
			Desc:   lipgloss.NewStyle().Foreground(lightWhite),
		},
		LogViewerStyles: LogViewerStyles{
			Text: lipgloss.NewStyle().Foreground(lightWhite),
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
						Selected: lipgloss.NewStyle().Foreground(colors.Black).Background(colors.White),
						Cell:     lipgloss.NewStyle().Foreground(lightWhite),
					},
				},
				FrameStyle:       defaultFrameStyle,
				MediaFilterStyle: lipgloss.NewStyle().Foreground(colors.Magenta1).Italic(true),
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
	Text  lipgloss.Style
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
