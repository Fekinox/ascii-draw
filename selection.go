package main

import "github.com/gdamore/tcell/v2"

type LassoTool struct {
	isLassoing  bool
	lassoPoints []Position
	topLeft     Position
	mask        Grid[bool]
}

func (l *LassoTool) HandleEvent(m *Editor, event tcell.Event) {
	switch ev := event.(type) {
	case *tcell.EventMouse:
		cx, cy := m.cursorX+m.sx, m.cursorY+m.sy
		if ev.Buttons()&tcell.Button1 != 0 {
			if !l.isLassoing {
				l.isLassoing = true
				l.lassoPoints = nil
			}
			p := Position{X: cx, Y: cy}
			if len(l.lassoPoints) == 0 || l.lassoPoints[len(l.lassoPoints)-1] != p {
				l.lassoPoints = append(l.lassoPoints, p)
			}
		} else if l.isLassoing {
			m.Stage()
			l.isLassoing = false
			topLeft, mask := CreateMask(l.lassoPoints)
			// convert from canvas to screen coords
			topLeft.X -= m.sx + m.offsetX
			topLeft.Y -= m.sy + m.offsetY
			m.stagingCanvas.SetSelection(mask, topLeft)
			m.Commit()
		}
	}
}

func (l *LassoTool) Draw(m *Editor, p Painter, x, y, w, h int, lag float64) {
	SetString(p, x+m.sx, y+m.sy-1, "Lasso Tool", tcell.StyleDefault)
	if l.isLassoing {
		j := len(l.lassoPoints) - 1
		for i, p1 := range l.lassoPoints {
			p2 := l.lassoPoints[j]
			points := LinePositions(p1.X, p1.Y, p2.X, p2.Y)
			for _, ps := range points {
				p.SetByte(ps.X, ps.Y, '#', tcell.StyleDefault)
			}
			j = i
		}
	}
}
