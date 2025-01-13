package main

import "github.com/gdamore/tcell/v2"

type StampTool struct {
	isDragging bool
	points     []Position
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
				l.points = nil
			}
			p := Position{X: cx, Y: cy}
			if len(l.points) == 0 || l.points[len(l.points)-1] != p {
				l.points = append(l.points, p)
			}
		} else if l.isDragging {
			l.isDragging = false
			m.Stage()
			m.stagingCanvas.Stamp(m.canvas, m.clipboard, l.points)
			m.Commit()
			l.points = nil
		}
	}
}

func (l *StampTool) Draw(m *MainWidget, p Painter, x, y, w, h int, lag float64) {
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
	if l.isDragging {
		for _, pt := range l.points {
			if m.canvas.Data.InBounds(pt.X, pt.Y) {
				// p.SetByte(
				// 	pt.X,
				// 	pt.Y,
				// 	m.brushCharacter,
				// 	tcell.StyleDefault.Foreground(m.fgColor).Background(m.bgColor),
				// )
				for y := range m.clipboard.Height {
					for x := range m.clipboard.Width {
						c := m.clipboard.MustGet(x, y)
						if c.Value != ' ' {
							crop.SetByte(
								pt.X+x+dx+m.sx+m.offsetX,
								pt.Y+y+dy+m.sy+m.offsetY,
								c.Value,
								c.Style,
							)
						}
					}
				}
			}
		}
	} else {
		for y := range m.clipboard.Height {
			for x := range m.clipboard.Width {
				c := m.clipboard.MustGet(x, y)
				if c.Value != ' ' {
					crop.SetByte(m.cursorX+x+dx, m.cursorY+y+dy, c.Value, c.Style)
				}
			}
		}
	}
}
