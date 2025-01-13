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
		cx, cy := m.cursorX+m.sx, m.cursorY+m.sy
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
			for _, pt := range LinePositions(l.origX, l.origY, m.cursorX+m.sx, m.cursorY+m.sy) {
				// m.canvas.Data.Set(pt.X-m.sx-m.offsetX, pt.Y-m.sy-m.offsetY, cell)
				m.canvas.FillRegion(
					pt.X-m.sx-m.offsetX-m.brushRadius/2,
					pt.Y-m.sy-m.offsetY-m.brushRadius/2,
					m.brushRadius,
					m.brushRadius,
					cell,
				)
			}
		}
	case *tcell.EventKey:
		if ev.Key() == tcell.KeyRune {
			m.brushCharacter = byte(ev.Rune())
		}
	}
}

func (l *LineTool) Draw(m *MainWidget, p Painter, x, y, w, h int, lag float64) {
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
		for _, pt := range LinePositions(l.origX, l.origY, m.cursorX+m.sx, m.cursorY+m.sy) {
			if m.canvas.Data.InBounds(pt.X-m.sx-m.offsetX, pt.Y-m.sy-m.offsetY) {
				// p.SetByte(
				// 	pt.X,
				// 	pt.Y,
				// 	m.brushCharacter,
				// 	tcell.StyleDefault.Foreground(m.fgColor).Background(m.bgColor),
				// )
				FillRegion(
					crop,
					pt.X-m.brushRadius/2, pt.Y-m.brushRadius/2,
					m.brushRadius, m.brushRadius,
					rune(m.brushCharacter),
					tcell.StyleDefault.Foreground(m.fgColor).Background(m.bgColor),
				)
			}
		}
	}
}
