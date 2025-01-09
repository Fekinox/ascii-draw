package main

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
)

var colorMap = map[rune]int{
	'1': 0, '!': 8,
	'2': 1, '@': 9,
	'3': 2, '#': 10,
	'4': 3, '$': 11,
	'5': 4, '%': 12,
	'6': 5, '^': 13,
	'7': 6, '&': 14,
	'8': 7, '*': 15,
	'0': 16,
}

type ColorSelectorTool struct {
	selFg int
}

var (
	_ Tool = &ColorSelectorTool{}
)

func MakeColorSelectorTool() *ColorSelectorTool {
	return &ColorSelectorTool{
		selFg: -1,
	}
}

func (c *ColorSelectorTool) HandleEvent(m *MainWidget, event tcell.Event) {
	switch ev := event.(type) {
	case *tcell.EventKey:
		if ev.Key() == tcell.KeyRune {
			r := ev.Rune()
			newColor, ok := colorMap[r]
			if !ok {
				m.ClearTool()
			}

			if c.selFg == -1 {
				c.selFg = newColor
				return
			}

			m.SetColor(c.selFg, newColor)
			m.ClearTool()
		}
	}
}

func (c *ColorSelectorTool) Update(m *MainWidget) {
}

func (c *ColorSelectorTool) Draw(
	m *MainWidget,
	p Painter,
	x, y, w, h int,
	lag float64,
) {
	if c.selFg == -1 {
		SetString(p, x+1, y, "fg: ", tcell.StyleDefault)
	} else {
		SetString(p, x+1, y, "bg: ", tcell.StyleDefault)
	}
	for i := range 8 {
		xx, yy := i%4, i/4
		color := tcell.Color(i) + tcell.ColorValid

		var st tcell.Style
		if c.selFg == -1 {
			st = st.Foreground(color)
		} else {
			st = st.Background(color).Foreground(tcell.ColorBlack)
		}

		SetString(p, x+5+xx*4, y+yy*2, fmt.Sprintf(" %d ", i+1), st)
	}

	for i := range 8 {
		xx, yy := i%4, i/4
		color := tcell.Color(i+8) + tcell.ColorValid

		var st tcell.Style
		if c.selFg == -1 {
			st = st.Foreground(color)
		} else {
			st = st.Background(color).Foreground(tcell.ColorBlack)
		}

		SetString(p, x+5+xx*4, y+yy*2+1, fmt.Sprintf("s+%d", i+1), st)
	}

	SetString(p, x+5+16, y, " 0 ", tcell.StyleDefault)
}
