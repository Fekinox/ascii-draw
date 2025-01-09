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

func (c *ColorPicker) Draw(
	m *MainWidget,
	p Painter,
	x, y, w, h int,
	lag float64,
) {
	if !c.inBounds {
		return
	}
}
