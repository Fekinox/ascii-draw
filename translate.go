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
			m.Stage()
			m.stagingCanvas.TranslateBlankTransparent(
				m.canvas,
				m.canvas.SelectionMask,
				Position{},
				cx-l.origX, cy-l.origY,
			)
		} else if l.isDragging {
			l.isDragging = false

			m.Commit()
		}
	}
}

func (l *TranslateTool) Draw(m *MainWidget, p Painter, x, y, w, h int, lag float64) {
}
