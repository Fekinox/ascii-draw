package main

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
)

var (
	Screen tcell.Screen
)

type Painter interface {
	SetByte(x, y int, v byte, style tcell.Style)
	SetRune(x, y int, v rune, combining []rune, style tcell.Style)
	SetStyle(x, y int, style tcell.Style)
	GetContent(x, y int) (rune, tcell.Style)
}

type DefaultPainter struct{}

type CropPainter struct {
	p            Painter
	offsetBefore Position
	area         Area
	offsetAfter  Position
}

var (
	Paint DefaultPainter
)

func (d DefaultPainter) SetByte(x, y int, v byte, style tcell.Style) {
	Screen.SetContent(x, y, rune(v), nil, style)
}

func (d DefaultPainter) SetRune(
	x, y int,
	v rune,
	combining []rune,
	style tcell.Style,
) {
	Screen.SetContent(x, y, rune(v), combining, style)
}

func (d DefaultPainter) GetContent(x, y int) (rune, tcell.Style) {
	rune, _, style, _ := Screen.GetContent(x, y)
	return rune, style
}

func (d DefaultPainter) SetStyle(x, y int, style tcell.Style) {
	pri, com, _, _ := Screen.GetContent(x, y)
	Screen.SetContent(x, y, pri, com, style)
}

func (a *CropPainter) SetByte(x, y int, v byte, style tcell.Style) {
	xx, yy := x+a.offsetBefore.X, y+a.offsetBefore.Y
	if a.area.Contains(xx, yy) {
		a.p.SetByte(xx+a.offsetAfter.X, yy+a.offsetAfter.Y, v, style)
	}
}

func (a *CropPainter) SetRune(
	x, y int,
	v rune,
	combining []rune,
	style tcell.Style,
) {
	xx, yy := x+a.offsetBefore.X, y+a.offsetBefore.Y
	if a.area.Contains(xx, yy) {
		a.p.SetRune(xx+a.offsetAfter.X, yy+a.offsetAfter.Y, v, combining, style)
	}
}

func (a *CropPainter) SetStyle(x, y int, style tcell.Style) {
	xx, yy := x+a.offsetBefore.X, y+a.offsetBefore.Y
	if a.area.Contains(xx, yy) {
		a.p.SetStyle(xx+a.offsetAfter.X, yy+a.offsetAfter.Y, style)
	}
}

func (a *CropPainter) GetContent(x, y int) (rune, tcell.Style) {
	xx, yy := x+a.offsetBefore.X, y+a.offsetBefore.Y
	if a.area.Contains(xx, yy) {
		rune, _, style, _ := Screen.GetContent(xx, yy)
		return rune, style
	}

	return 0, tcell.StyleDefault
}

type Span struct {
	Contents string
	Style    tcell.Style
}

func SetString(p Painter, x int, y int, s string, style tcell.Style) {
	col := x
	for _, ch := range s {
		width := runewidth.RuneWidth(ch)
		p.SetRune(col, y, ch, nil, style)
		col += width
	}
}

func SetCenteredString(p Painter, x, y int, s string, style tcell.Style) {
	col := x - runewidth.StringWidth(s)/2
	for _, ch := range s {
		width := runewidth.RuneWidth(ch)
		p.SetRune(col, y, ch, nil, style)
		col += width
	}
}

func SetStringArray(
	p Painter,
	x, y int,
	style tcell.Style, leftAlign bool,
	strings ...string) {

	for i, s := range strings {
		xx := x
		if leftAlign {
			xx -= runewidth.StringWidth(s)
		}

		SetString(p, xx, y+i, s, style)
	}
}

func SetCenteredSpans(p Painter, x, y int, spans ...Span) {
	width := 0
	for _, sp := range spans {
		width += runewidth.StringWidth(sp.Contents)
	}

	col := x - width/2
	for _, sp := range spans {
		SetString(p, col, y, sp.Contents, sp.Style)
		col += runewidth.StringWidth(sp.Contents)
	}
}

func SetGrid(p Painter, x, y int, grid Grid[rune], style tcell.Style) {
	for dy := 0; dy < grid.Height; dy++ {
		for dx := 0; dx < grid.Width; dx++ {
			p.SetRune(
				x+dx,
				y+dy,
				grid.MustGet(dx, dy),
				nil, style)
		}
	}
}

func FillRegion(
	p Painter,
	x, y int,
	width, height int,
	c rune,
	style tcell.Style,
) {
	for dy := 0; dy < height; dy++ {
		for dx := 0; dx < width; dx++ {
			p.SetRune(x+dx, y+dy, c, nil, style)
		}
	}
}

func ShowResizeScreen(p Painter, w, h int, style tcell.Style) {
	SetCenteredString(p, w/2, h/2, "Screen too small!", style)
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

	SetCenteredSpans(p, w/2, h/2+1,
		Span{Contents: "Current: ", Style: style},
		widthSpan,
		Span{Contents: " x ", Style: style},
		heightSpan,
	)
}

func BorderBox(p Painter, area Area, style tcell.Style) {
	// Draw corners
	p.SetRune(area.X, area.Y, tcell.RuneULCorner, nil, style)
	p.SetRune(area.X+area.Width-1, area.Y, tcell.RuneURCorner, nil, style)
	p.SetRune(
		area.X,
		area.Y+area.Height-1,
		tcell.RuneLLCorner,
		nil,
		style,
	)
	p.SetRune(
		area.X+area.Width-1,
		area.Y+area.Height-1,
		tcell.RuneLRCorner,
		nil,
		style,
	)

	// Draw top and bottom edges
	for xx := area.X + 1; xx < area.X+area.Width-1; xx++ {
		p.SetRune(xx, area.Y, tcell.RuneHLine, nil, style)
		p.SetRune(xx, area.Y+area.Height-1, tcell.RuneHLine, nil, style)
	}

	// Draw left and right edges
	for yy := area.Y + 1; yy < area.Y+area.Height-1; yy++ {
		p.SetRune(area.X, yy, tcell.RuneVLine, nil, style)
		p.SetRune(area.X+area.Width-1, yy, tcell.RuneVLine, nil, style)
	}
}

// Bresenham's line algorithm
func drawLineLow(p Painter, ax, ay, bx, by int) {
	dx, dy := bx-ax, by-ay
	yi := 1
	if dy < 0 {
		yi = -1
		dy = -dy
	}
	D := 2*dy - dx
	y := ay

	for x := ax; x <= bx; x++ {
		p.SetRune(x, y, '#', nil, tcell.StyleDefault)
		if D > 0 {
			y += yi
			D += 2 * (dy - dx)
		} else {
			D += 2 * dy
		}
	}
}

func drawLineHigh(p Painter, ax, ay, bx, by int) {
	dx, dy := bx-ax, by-ay
	xi := 1
	if dx < 0 {
		xi = -1
		dx = -dx
	}
	D := 2*dx - dy
	x := ax

	for y := ay; y <= by; y++ {
		p.SetRune(x, y, '#', nil, tcell.StyleDefault)
		if D > 0 {
			x += xi
			D += 2 * (dx - dy)
		} else {
			D += 2 * dx
		}
	}
}

func DrawLine(p Painter, ax, ay, bx, by int) {
	dx, dy := bx-ax, by-ay
	if max(dy, -dy) < max(dx, -dx) {
		if dx < 0 {
			ax, ay, bx, by = bx, by, ax, ay
		}
		drawLineLow(p, ax, ay, bx, by)
	} else {
		if dy < 0 {
			ax, ay, bx, by = bx, by, ax, ay
		}
		drawLineHigh(p, ax, ay, bx, by)
	}
}

func DrawColorPickerState(
	p Painter,
	x, y int,
	copyChar, copyFg, copyBg bool,
	hoverChar byte, hoverFg, hoverBg tcell.Color,
) {
	rect := Area{
		X:      x,
		Y:      y - 5,
		Width:  8,
		Height: 5,
	}

	// char
	p.SetByte(x+1, y-4, hoverChar, tcell.StyleDefault)
	SetString(p, x+2, y-4, " char  ", tcell.StyleDefault)

	// fg
	if hoverFg == 0 {
		p.SetByte(x+1, y-3, '_', tcell.StyleDefault)
	} else {
		var ch byte = 'b'
		if hoverFg <= tcell.ColorGray {
			ch = 'n'
		}
		p.SetByte(x+1, y-3, ch, tcell.StyleDefault.Background(hoverFg))
	}
	SetString(p, x+2, y-3, " fg    ", tcell.StyleDefault)

	// bg
	if hoverBg == 0 {
		p.SetByte(x+1, y-2, '_', tcell.StyleDefault)
	} else {
		var ch byte = 'b'
		if hoverBg <= tcell.ColorGray {
			ch = 'n'
		}
		p.SetByte(x+1, y-2, ch, tcell.StyleDefault.Background(hoverBg))
	}
	SetString(p, x+2, y-2, " bg    ", tcell.StyleDefault)

	BorderBox(p, rect, tcell.StyleDefault)
}

func DrawDragIndicator(
	p Painter,
	x, y, orx, ory int,
) {
	origRect := Area{
		X:      orx - 3,
		Y:      ory - 2,
		Width:  7,
		Height: 5,
	}
	indRect := Area{
		X:      x + 2,
		Y:      y - 1,
		Width:  10,
		Height: 3,
	}
	offset := y - ory

	var s string
	if offset < -2 {
		s = "set fg"
	} else if offset > 2 {
		s = "set bg"
	} else {
		s = "set char"
	}

	BorderBox(p, origRect, tcell.StyleDefault)
	SetString(p, x+3, y, s, tcell.StyleDefault)
	BorderBox(p, indRect, tcell.StyleDefault)
}
