package main

import "github.com/gdamore/tcell/v2"

type TranslateTool struct {
	isDragging bool
	origX      int
	origY      int
}

func (l *TranslateTool) HandleEvent(m *MainWidget, event tcell.Event) {
	switch ev := event.(type) {
	case *tcell.EventMouse:
		cx, cy := m.cursorX+m.sx, m.cursorY+m.sy
		if ev.Buttons()&tcell.Button1 != 0 {
			if !l.isDragging {
				l.isDragging = true
				l.origX, l.origY = cx, cy
			}
			m.SetTransform(cx-l.origX, cy-l.origY)
		} else if l.isDragging {
			l.isDragging = false

			m.SetTransform(cx-l.origX, cy-l.origY)
			m.CommitTransform()
		}
	}
}

func (l *TranslateTool) Draw(m *MainWidget, p Painter, x, y, w, h int, lag float64) {
}
