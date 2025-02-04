package main

import (
	"time"

	"github.com/gdamore/tcell/v2"
)

type BrushTool struct {
	isDragging bool

	lineMode bool
	start    Position

	lastPaint    time.Time
	lastPaintPos Position
}

func (b *BrushTool) HandleEvent(m *Editor, event tcell.Event) {
	switch ev := event.(type) {
	case *tcell.EventMouse:
		cx, cy := m.cursorX-m.brushRadius/2-m.offsetX, m.cursorY-m.brushRadius/2-m.offsetY
		p := Position{X: cx, Y: cy}
		cell := Cell{
			Value: m.brushCharacter,
			Style: tcell.StyleDefault.Foreground(m.fgColor).Background(m.bgColor),
		}

		if !b.lineMode {
			if ev.Buttons()&tcell.Button1 != 0 {
				if !b.isDragging {
					b.isDragging = true
				}

				m.Stage()
				cell := Cell{
					Value: m.brushCharacter,
					Style: tcell.StyleDefault.Foreground(m.fgColor).Background(m.bgColor),
				}

				if ev.When().Sub(b.lastPaint).Seconds() < 0.1 {
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
			} else if b.isDragging {
				b.isDragging = false
				m.Commit()
			}
		} else {
			if ev.Buttons()&tcell.Button1 != 0 {
				if !b.isDragging {
					b.isDragging = true
					b.start = Position{X: cx, Y: cy}
				}
			} else if b.isDragging {
				b.isDragging = false
				m.Stage()
				linePositions := LinePositions(b.start.X, b.start.Y, cx, cy)
				m.stagingCanvas.BrushStrokes(m.brushRadius, cell, linePositions, m.lockMask)
				m.Commit()
			}
		}
		b.lastPaintPos = p
		b.lastPaint = ev.When()
	case *tcell.EventKey:
		if ev.Key() == tcell.KeyTab {
			if b.lineMode {
				m.Stage()
				cell := Cell{
					Value: m.brushCharacter,
					Style: tcell.StyleDefault.Foreground(m.fgColor).Background(m.bgColor),
				}
				linePositions := LinePositions(b.start.X, b.start.Y, b.lastPaintPos.X, b.lastPaintPos.Y)
				m.stagingCanvas.BrushStrokes(m.brushRadius, cell, linePositions, m.lockMask)
				m.Commit()
			} else {
				m.Commit()
			}
			b.isDragging = false
			b.lineMode = !b.lineMode
		}
	}
}

func (b *BrushTool) Draw(m *Editor, p Painter, x, y, w, h int, lag float64) {
	if b.lineMode {
		SetString(p, x+m.sx, y+m.sy-1, "Brush Tool (straight lines)", tcell.StyleDefault)
	} else {
		SetString(p, x+m.sx, y+m.sy-1, "Brush Tool", tcell.StyleDefault)
	}
	if b.isDragging && b.lineMode {
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
			b.start.X, b.start.Y,
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
