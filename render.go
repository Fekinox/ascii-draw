package main

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
)

var (
	Screen tcell.Screen
)

type Span struct {
	Contents string
	Style    tcell.Style
}

func SetString(x int, y int, s string, style tcell.Style) {
	col := x
	for _, ch := range s {
		width := runewidth.RuneWidth(ch)
		Screen.SetContent(col, y, ch, nil, style)
		col += width
	}
}

func SetCenteredString(x, y int, s string, style tcell.Style) {
	col := x - runewidth.StringWidth(s)/2
	for _, ch := range s {
		width := runewidth.RuneWidth(ch)
		Screen.SetContent(col, y, ch, nil, style)
		col += width
	}
}

func SetStringArray(
	x, y int,
	style tcell.Style, leftAlign bool,
	strings ...string) {

	for i, s := range strings {
		xx := x
		if leftAlign {
			xx -= runewidth.StringWidth(s)
		}

		SetString(xx, y+i, s, style)
	}
}

func SetCenteredSpans(x, y int, spans ...Span) {
	width := 0
	for _, sp := range spans {
		width += runewidth.StringWidth(sp.Contents)
	}

	col := x - width/2
	for _, sp := range spans {
		SetString(col, y, sp.Contents, sp.Style)
		col += runewidth.StringWidth(sp.Contents)
	}
}

func SetGrid(x, y int, grid Grid[rune], style tcell.Style) {
	for dy := 0; dy < grid.Height; dy++ {
		for dx := 0; dx < grid.Width; dx++ {
			Screen.SetContent(
				x+dx,
				y+dy,
				grid.MustGet(dx, dy),
				nil, style)
		}
	}
}

func FillRegion(x, y int, width, height int, c rune, style tcell.Style) {
	for dy := 0; dy < height; dy++ {
		for dx := 0; dx < width; dx++ {
			Screen.SetContent(x+dx, y+dy, c, nil, style)
		}
	}
}

func ShowResizeScreen(w, h int, style tcell.Style) {
	SetCenteredString(w/2, h/2, "Screen too small!", style)
	var widthColor, heightColor tcell.Color
	if w < MIN_WIDTH {
		widthColor = tcell.ColorRed
	} else {
		widthColor = tcell.ColorGreen
	}
	if h < MIN_HEIGHT {
		heightColor = tcell.ColorRed
	} else {
		heightColor = tcell.ColorGreen
	}

	widthSpan := Span{
		Contents: fmt.Sprintf("%d", w),
		Style:    style.Bold(true).Foreground(widthColor),
	}
	heightSpan := Span{
		Contents: fmt.Sprintf("%d", h),
		Style:    style.Bold(true).Foreground(heightColor),
	}

	SetCenteredSpans(w/2, h/2+1,
		Span{Contents: "Current: ", Style: style},
		widthSpan,
		Span{Contents: " x ", Style: style},
		heightSpan,
	)
}

func BorderBox(area Area, style tcell.Style) {
	// Draw corners
	Screen.SetContent(area.X, area.Y, tcell.RuneULCorner, nil, style)
	Screen.SetContent(area.X+area.Width, area.Y, tcell.RuneURCorner, nil, style)
	Screen.SetContent(
		area.X,
		area.Y+area.Height,
		tcell.RuneLLCorner,
		nil,
		style,
	)
	Screen.SetContent(
		area.X+area.Width,
		area.Y+area.Height,
		tcell.RuneLRCorner,
		nil,
		style,
	)

	// Draw top and bottom edges
	for xx := area.X + 1; xx < area.X+area.Width; xx++ {
		Screen.SetContent(xx, area.Y, tcell.RuneHLine, nil, style)
		Screen.SetContent(xx, area.Y+area.Height, tcell.RuneHLine, nil, style)
	}

	// Draw left and right edges
	for yy := area.Y + 1; yy < area.Y+area.Height; yy++ {
		Screen.SetContent(area.X, yy, tcell.RuneVLine, nil, style)
		Screen.SetContent(area.X+area.Width, yy, tcell.RuneVLine, nil, style)
	}
}

// Bresenham's line algorithm
func drawLineLow(screen tcell.Screen, ax, ay, bx, by int) {
	dx, dy := bx-ax, by-ay
	yi := 1
	if dy < 0 {
		yi = -1
		dy = -dy
	}
	D := 2*dy - dx
	y := ay

	for x := ax; x <= bx; x++ {
		screen.SetContent(x, y, '#', nil, tcell.StyleDefault)
		if D > 0 {
			y += yi
			D += 2 * (dy - dx)
		} else {
			D += 2 * dy
		}
	}
}

func drawLineHigh(screen tcell.Screen, ax, ay, bx, by int) {
	dx, dy := bx-ax, by-ay
	xi := 1
	if dx < 0 {
		xi = -1
		dx = -dx
	}
	D := 2*dx - dy
	x := ax

	for y := ay; y <= by; y++ {
		screen.SetContent(x, y, '#', nil, tcell.StyleDefault)
		if D > 0 {
			x += xi
			D += 2 * (dx - dy)
		} else {
			D += 2 * dx
		}
	}
}

func DrawLine(screen tcell.Screen, ax, ay, bx, by int) {
	dx, dy := bx-ax, by-ay
	if max(dy, -dy) < max(dx, -dx) {
		if dx < 0 {
			ax, ay, bx, by = bx, by, ax, ay
		}
		drawLineLow(screen, ax, ay, bx, by)
	} else {
		if dy < 0 {
			ax, ay, bx, by = bx, by, ax, ay
		}
		drawLineHigh(screen, ax, ay, bx, by)
	}
}
