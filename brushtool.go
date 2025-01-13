package main

import (
	"time"

	"github.com/gdamore/tcell/v2"
)

type BrushTool struct {
	isDragging bool
	points     []Position
	lastPaint  time.Time
}

func (b *BrushTool) HandleEvent(m *MainWidget, event tcell.Event) {
	switch ev := event.(type) {
	case *tcell.EventMouse:
		cx, cy := m.cursorX+m.sx, m.cursorY+m.sy
		if ev.Buttons()&tcell.Button1 != 0 {
			if !b.isDragging {
				b.isDragging = true
				b.points = nil
			}
			p := Position{X: cx, Y: cy}
			if len(b.points) == 0 {
				b.points = append(b.points, p)
			} else if last := b.points[len(b.points)-1]; last != p {
				dx, dy := cx-last.X, cy-last.Y
				dist := max(max(dx, -dx), max(dy, -dy))
				if dist > 1 && ev.When().Sub(b.lastPaint).Seconds() < 0.01 {
					posns := LinePositions(last.X, last.Y, cx, cy)
					b.points = append(b.points, posns[1:len(posns)-1]...)
					b.points = append(b.points, p)
				}
				b.points = append(b.points, p)
			}
			b.lastPaint = ev.When()
		} else if b.isDragging {
			b.isDragging = false
			cell := Cell{
				Value: m.brushCharacter,
				Style: tcell.StyleDefault.Foreground(m.fgColor).Background(m.bgColor),
			}
			m.Stage()
			for _, p := range b.points {
				// m.canvas.Data.Set(pt.X-m.sx-m.offsetX, pt.Y-m.sy-m.offsetY, cell)
				m.stagingCanvas.FillRegion(
					p.X-m.sx-m.offsetX-m.brushRadius/2,
					p.Y-m.sy-m.offsetY-m.brushRadius/2,
					m.brushRadius,
					m.brushRadius,
					cell,
				)
			}
			m.Commit()
			b.points = nil
		}
	case *tcell.EventKey:
		if ev.Key() == tcell.KeyRune {
			m.brushCharacter = byte(ev.Rune())
		}
	}
}

func (b *BrushTool) Draw(m *MainWidget, p Painter, x, y, w, h int, lag float64) {
	if !b.isDragging {
		return
	}

	crop := &CropPainter{
		p: p,
		area: Area{
			X:      m.offsetX + m.sx,
			Y:      m.offsetY + m.sy,
			Width:  m.canvas.Data.Width,
			Height: m.canvas.Data.Height,
		},
	}
	for _, pt := range b.points {
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
