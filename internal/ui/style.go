package ui

import (
	"github.com/clambin/videoConvertor/internal/worklist"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const (
	labelColor    = tcell.ColorSandyBrown
	textColor     = tcell.ColorGrey
	shortcutColor = tcell.ColorBlue
)

var tableColorStatus = map[worklist.WorkStatus]tcell.Color{
	worklist.Failed:     tcell.ColorRed,
	worklist.Rejected:   tcell.ColorRed,
	worklist.Converted:  tcell.ColorGreen,
	worklist.Skipped:    tcell.ColorGreen,
	worklist.Converting: tcell.ColorOrange,
}

func init() {
	tview.Styles.TitleColor = tcell.ColorMediumTurquoise
	tview.Styles.BorderColor = tcell.ColorLightSkyBlue
	tview.Styles.PrimaryTextColor = tcell.ColorLightSkyBlue
	tview.Styles.SecondaryTextColor = tcell.ColorWhite
	tview.Styles.TertiaryTextColor = tcell.ColorGrey
}
