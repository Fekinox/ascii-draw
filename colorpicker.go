package main

import "github.com/gdamore/tcell/v2"

type ColorPicker struct {
	copyChar bool
	copyFg   bool
	copyBg   bool

	inBounds  bool
	hoverChar byte
	hoverFg   tcell.Color
	hoverBg   tcell.Color
}

var (
	_ Tool = &ColorPicker{}
)

func MakeColorPickerTool() *ColorPicker {
	return &ColorPicker{
		copyChar: true,
		copyFg:   true,
	}
}

func (c *ColorPicker) HandleEvent(m *MainWidget, event tcell.Event) {
	switch ev := event.(type) {
	case *tcell.EventMouse:
		cx, cy := m.cursorX+m.sx, m.cursorY+m.sy

		canvasX, canvasY := cx-m.sx-m.offsetX, cy-m.sy-m.offsetY
		if m.canvas.Data.InBounds(canvasX, canvasY) {
			c.inBounds = true
			cell := m.canvas.Data.MustGet(canvasX, canvasY)
			c.hoverChar = cell.Value
			c.hoverFg, c.hoverBg, _ = cell.Style.Decompose()
		} else {
			c.inBounds = false
		}

		if ev.Buttons()&tcell.Button1 != 0 {
			if c.copyChar {
				m.brushCharacter = c.hoverChar
			}

			if c.copyFg {
				m.fgColor = c.hoverFg
			}

			if c.copyBg {
				m.bgColor = c.hoverBg
			}
		}
	case *tcell.EventKey:
		if ev.Key() == tcell.KeyRune {
			switch ev.Rune() {
			case '1':
				c.copyChar = !c.copyChar
			case '2':
				c.copyFg = !c.copyFg
			case '3':
				c.copyBg = !c.copyBg
			}
		}
	}
}

func (c *ColorPicker) Draw(m *MainWidget, p Painter, x, y, w, h int, lag float64) {
	if !c.inBounds {
		return
	}

	cx, cy := m.cursorX+m.sx, m.cursorY+m.sy

	rect := Area{
		X:      cx,
		Y:      cy - 5,
		Width:  10,
		Height: 5,
	}

	// char
	p.SetByte(cx+1, cy-4, c.hoverChar, tcell.StyleDefault)
	SetString(p, cx+2, cy-4, " char  ", tcell.StyleDefault)
	// fg
	if c.hoverFg == 0 {
		p.SetByte(cx+1, cy-3, '_', tcell.StyleDefault)
	} else {
		var ch byte = 'b'
		if m.fgColor <= tcell.ColorGray {
			ch = 'n'
		}
		p.SetByte(cx+1, cy-3, ch, tcell.StyleDefault.Background(c.hoverFg))
	}
	SetString(p, cx+2, cy-3, " fg    ", tcell.StyleDefault)
	// bg
	if c.hoverBg == 0 {
		p.SetByte(cx+1, cy-2, '_', tcell.StyleDefault)
	} else {
		var ch byte = 'b'
		if m.bgColor <= tcell.ColorGray {
			ch = 'n'
		}
		p.SetByte(cx+1, cy-2, ch, tcell.StyleDefault.Background(c.hoverBg))
	}
	SetString(p, cx+2, cy-2, " bg    ", tcell.StyleDefault)

	if c.copyChar {
		p.SetByte(cx+8, cy-4, '+', tcell.StyleDefault)
	}

	if c.copyFg {
		p.SetByte(cx+8, cy-3, '+', tcell.StyleDefault)
	}

	if c.copyBg {
		p.SetByte(cx+8, cy-2, '+', tcell.StyleDefault)
	}

	BorderBox(p, rect, tcell.StyleDefault)
}
