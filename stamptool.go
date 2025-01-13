package main

import "github.com/gdamore/tcell/v2"

type StampTool struct{}

func (l *StampTool) HandleEvent(m *MainWidget, event tcell.Event) {
	switch ev := event.(type) {
	case *tcell.EventMouse:
		cx, cy := m.cursorX-m.offsetX, m.cursorY-m.offsetY
		dx, dy := -m.clipboard.Width/2, -m.clipboard.Height/2
		if ev.Buttons()&tcell.Button1 != 0 {
			if m.clipboard.Width != 0 && m.clipboard.Height != 0 {
				for y := range m.clipboard.Height {
					for x := range m.clipboard.Width {
						c := m.clipboard.MustGet(x, y)
						if c.Value != ' ' {
							m.canvas.Data.Set(x+cx+dx, y+cy+dy, c)
						}
					}
				}
			}
		}
	}
}

func (l *StampTool) Draw(m *MainWidget, p Painter, x, y, w, h int, lag float64) {
	dx, dy := -m.clipboard.Width/2, -m.clipboard.Height/2
	px, py := m.cursorX+dx+m.sx, m.cursorY+dy+m.sy

	if m.clipboard.Width != 0 && m.clipboard.Height != 0 {
		for y := range m.clipboard.Height {
			for x := range m.clipboard.Width {
				c := m.clipboard.MustGet(x, y)
				if c.Value != ' ' {
					p.SetByte(px+x, py+y, c.Value, c.Style)
				}
			}
		}
	}
}
