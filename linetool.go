package main

import "github.com/gdamore/tcell/v2"

type LineTool struct {
	isDragging bool
	origX      int
	origY      int
}

func (l *LineTool) HandleEvent(m *MainWidget, event tcell.Event) {
	switch ev := event.(type) {
	case *tcell.EventMouse:
		cx, cy := m.cursorX-m.offsetX, m.cursorY-m.offsetY
		if ev.Buttons()&tcell.Button1 != 0 {
			if !l.isDragging {
				l.isDragging = true
				l.origX, l.origY = cx, cy
			}
		} else if l.isDragging {
			l.isDragging = false
			cell := Cell{
				Value: m.brushCharacter,
				Style: tcell.StyleDefault.Foreground(m.fgColor).Background(m.bgColor),
			}
			m.Stage()
			linePositions := LinePositions(l.origX, l.origY, cx, cy)
			m.stagingCanvas.BrushStrokes(m.brushRadius, cell, linePositions, m.lockMask)
			m.Commit()
		}
	case *tcell.EventKey:
		if ev.Key() == tcell.KeyRune {
			m.brushCharacter = byte(ev.Rune())
		}
	}
}

func (l *LineTool) Draw(m *MainWidget, p Painter, x, y, w, h int, lag float64) {
	SetString(p, x+m.sx, y+m.sy-1, "Line Tool", tcell.StyleDefault)
	if l.isDragging {
		crop := &CropPainter{
			p: p,
			area: Area{
				X:      m.offsetX + m.sx,
				Y:      m.offsetY + m.sy,
				Width:  m.canvas.Data.Width,
				Height: m.canvas.Data.Height,
			},
		}
		for _, pt := range LinePositions(
			l.origX, l.origY,
			m.cursorX-m.offsetX, m.cursorY-m.offsetY) {
			if m.canvas.Data.InBounds(pt.X, pt.Y) {
				FillRegion(
					crop,
					pt.X-m.brushRadius/2+m.offsetX+m.sx, pt.Y-m.brushRadius/2+m.offsetY+m.sy,
					m.brushRadius, m.brushRadius,
					rune(m.brushCharacter),
					tcell.StyleDefault.Foreground(m.fgColor).Background(m.bgColor),
				)
			}
		}
	}
}
