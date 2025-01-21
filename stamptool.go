package main

import "github.com/gdamore/tcell/v2"

type StampTool struct {
	isDragging   bool
	hasLastPaint bool
	lastPaintPos Position
}

func (l *StampTool) HandleEvent(m *MainWidget, event tcell.Event) {
	if m.clipboard.Width == 0 || m.clipboard.Height == 0 {
		return
	}

	switch ev := event.(type) {
	case *tcell.EventMouse:
		cx, cy := m.cursorX-m.offsetX-m.sx, m.cursorY-m.offsetY-m.sy
		if ev.Buttons()&tcell.Button1 != 0 {
			if !l.isDragging {
				l.isDragging = true
				l.hasLastPaint = false
			}

			m.Stage()
			p := Position{X: cx, Y: cy}
			if !l.hasLastPaint || l.lastPaintPos != p {
				m.stagingCanvas.Stamp(m.clipboard, p.X, p.Y, m.lockMask)
			}
			l.hasLastPaint = true
			l.lastPaintPos = p
		} else if l.isDragging {
			l.isDragging = false
			m.Commit()
		}
	}
}

func (l *StampTool) Draw(m *MainWidget, p Painter, x, y, w, h int, lag float64) {
	SetString(p, x+m.sx, y+m.sy-1, "Stamp Tool", tcell.StyleDefault)
	if m.clipboard.Width == 0 || m.clipboard.Height == 0 {
		return
	}
	dx, dy := -m.clipboard.Width/2, -m.clipboard.Height/2

	crop := &CropPainter{
		p: p,
		area: Area{
			X:      m.offsetX + m.sx,
			Y:      m.offsetY + m.sy,
			Width:  m.canvas.Data.Width,
			Height: m.canvas.Data.Height,
		},
	}
	for y := range m.clipboard.Height {
		for x := range m.clipboard.Width {
			c := m.clipboard.MustGet(x, y)
			if c.Value != ' ' {
				crop.SetByte(m.cursorX+x+dx, m.cursorY+y+dy, c.Value, c.Style)
			}
		}
	}
}
