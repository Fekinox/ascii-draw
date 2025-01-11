package main

import "github.com/gdamore/tcell/v2"

type LassoTool struct {
	isLassoing  bool
	lassoPoints []Position
	topLeft     Position
	mask        Grid[bool]
}

func (l *LassoTool) HandleEvent(m *MainWidget, event tcell.Event) {
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

			l.topLeft, l.mask = CreateMask(l.lassoPoints)
		} else {
			l.isLassoing = false
		}
	}
}

func (l *LassoTool) Draw(m *MainWidget, p Painter, x, y, w, h int, lag float64) {
	for y := range l.mask.Height {
		for x := range l.mask.Width {
			if !l.mask.MustGet(x, y) {
				continue
			}
			xx, yy := x+l.topLeft.X, y+l.topLeft.Y
			_, s := p.GetContent(xx, yy)
			p.SetStyle(xx, yy, s.Reverse(true))
		}
	}
	// j := len(l.lassoPoints) - 1
	// for i, p1 := range l.lassoPoints {
	// 	p2 := l.lassoPoints[j]
	// 	points := LinePositions(p1.X, p1.Y, p2.X, p2.Y)
	// 	for _, ps := range points {
	// 		_, s := p.GetContent(ps.X, ps.Y)
	// 		p.SetStyle(ps.X, ps.Y, s.Reverse(true))
	// 	}
	// 	j = i
	// }
}
