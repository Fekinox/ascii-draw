package main

import (
	"math"
	"slices"

	"github.com/gdamore/tcell/v2"
)

type EditorState interface {
	OnEnter(ctx *MainWidget)
	HandleAction(ctx *MainWidget, action Action)
	HandleEvent(ctx *MainWidget, event tcell.Event)
	Update(ctx *MainWidget)
	Draw(ctx *MainWidget, screen tcell.Screen, x, y, w, h int, lag float64)
}

type NormalState struct {
}

func (n *NormalState) OnEnter(ctx *MainWidget) {
}

func (n *NormalState) HandleAction(ctx *MainWidget, action Action) {
}

func (n *NormalState) HandleEvent(ctx *MainWidget, event tcell.Event) {
	switch ev := event.(type) {
	case *tcell.EventMouse:
		cx, cy := ev.Position()

		if ev.Buttons()&tcell.Button1 != 0 {
			st := &LassoState{}
			ctx.SetState(st)
			st.lassoPoints = []Position{{X: cx, Y: cy}}
		}

	}
}

func (n *NormalState) Update(ctx *MainWidget) {
}

func (n *NormalState) Draw(ctx *MainWidget, screen tcell.Screen, x, y, w, h int, lag float64) {
}

type LassoState struct {
	lassoPoints []Position
}

func (l *LassoState) OnEnter(ctx *MainWidget) {
	l.lassoPoints = nil
}

func (l *LassoState) HandleAction(ctx *MainWidget, action Action) {
}

func (l *LassoState) HandleEvent(ctx *MainWidget, event tcell.Event) {
	switch ev := event.(type) {
	case *tcell.EventMouse:
		cx, cy := ev.Position()

		if ev.Buttons()&tcell.Button1 != 0 {
			p := Position{X: cx, Y: cy}
			if len(l.lassoPoints) == 0 || l.lassoPoints[len(l.lassoPoints)-1] != p {
				l.lassoPoints = append(l.lassoPoints, p)
				ctx.updateSelectionMask(l.lassoPoints)
			}
		} else {
			st := &NormalState{}
			ctx.SetState(st)
		}
	}
}

func (l *LassoState) Update(ctx *MainWidget) {
}

func (l *LassoState) Draw(ctx *MainWidget, screen tcell.Screen, x, y, w, h int, lag float64) {
	j := len(l.lassoPoints) - 1
	for i, p1 := range l.lassoPoints {
		p2 := l.lassoPoints[j]
		DrawLine(screen, p1.X, p1.Y, p2.X, p2.Y)
		j = i
	}
}

func (m *MainWidget) updateSelectionMask(lassoPoints []Position) {
	// Darel Rex Finley polygon filling
	var nodes int
	nodeX := make([]int, len(lassoPoints))

	for y := range m.selectionMask.Height {
		for x := range m.selectionMask.Width {
			m.selectionMask.Set(x, y, false)
		}
	}
	for yy := range m.selectionMask.Height {
		nodes = 0
		j := len(lassoPoints) - 1
		for i := 0; i < len(lassoPoints); i++ {
			p1, p2 := lassoPoints[i], lassoPoints[j]
			if p1.Y < yy && p2.Y >= yy ||
				p2.Y < yy && p1.Y >= yy {
				t := (float64(yy) - float64(p1.Y)) / (float64(p2.Y) - float64(p1.Y))
				nx := math.Round(float64(p1.X) + t*float64(p2.X-p1.X))
				nodeX[nodes] = int(nx)
				nodes++
			}
			j = i
		}

		slices.Sort(nodeX[:nodes])
		for i := 0; i < nodes; i += 2 {
			if nodeX[i] >= m.selectionMask.Width-1 {
				break
			}
			if nodeX[i+1] > 0 {
				if nodeX[i] < 0 {
					nodeX[i] = 0
				}
				if nodeX[i+1] > m.selectionMask.Width-1 {
					nodeX[i+1] = m.selectionMask.Width - 1
				}
				for xx := nodeX[i]; xx < nodeX[i+1]; xx++ {
					m.selectionMask.Set(xx, yy, true)
				}
			}
		}
	}
}
