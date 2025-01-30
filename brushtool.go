package main

import (
	"time"

	"github.com/gdamore/tcell/v2"
)

type BrushTool struct {
	isDragging   bool
	lastPaint    time.Time
	lastPaintPos Position
}

func (b *BrushTool) HandleEvent(m *Editor, event tcell.Event) {
	switch ev := event.(type) {
	case *tcell.EventMouse:
		cx, cy := m.cursorX-m.brushRadius/2-m.offsetX, m.cursorY-m.brushRadius/2-m.offsetY
		if ev.Buttons()&tcell.Button1 != 0 {
			if !b.isDragging {
				b.isDragging = true
			}

			m.Stage()
			cell := Cell{
				Value: m.brushCharacter,
				Style: tcell.StyleDefault.Foreground(m.fgColor).Background(m.bgColor),
			}
			p := Position{X: cx, Y: cy}

			if ev.When().Sub(b.lastPaint).Seconds() < 0.01 {
				dx, dy := cx-b.lastPaintPos.X, cy-b.lastPaintPos.Y
				dist := max(max(dx, -dx), max(dy, -dy))
				if dist > 1 {
					posns := LinePositions(b.lastPaintPos.X, b.lastPaintPos.Y, cx, cy)
					for _, pt := range posns[1 : len(posns)-1] {
						m.stagingCanvas.FillRegion(
							pt.X, pt.Y, m.brushRadius, m.brushRadius, cell, m.lockMask,
						)
					}
				}
				m.stagingCanvas.FillRegion(
					p.X, p.Y, m.brushRadius, m.brushRadius, cell, m.lockMask,
				)
			} else {
				m.stagingCanvas.FillRegion(
					p.X, p.Y, m.brushRadius, m.brushRadius, cell, m.lockMask,
				)
			}
			b.lastPaintPos = p
			b.lastPaint = ev.When()
		} else if b.isDragging {
			b.isDragging = false
			m.Commit()
		}
	case *tcell.EventKey:
		if ev.Key() == tcell.KeyRune {
			m.brushCharacter = byte(ev.Rune())
		}
	}
}

func (b *BrushTool) Draw(m *Editor, p Painter, x, y, w, h int, lag float64) {
	SetString(p, x+m.sx, y+m.sy-1, "Brush Tool", tcell.StyleDefault)
}
