package main

import (
	"time"

	"github.com/gdamore/tcell/v2"
)

type BrushTool struct {
	currentIcon  byte
	lastPaint    time.Time
	lastPosition Position
}

var (
	_ Tool = &BrushTool{}
)

func (b *BrushTool) Name() string { return "Brush" }

func (b *BrushTool) HandleAction(m *MainWidget, action Action) {
}

func (b *BrushTool) HandleEvent(m *MainWidget, event tcell.Event) {
	switch ev := event.(type) {
	case *tcell.EventKey:
		if ev.Key() == tcell.KeyRune {
			b.currentIcon = byte(ev.Rune())
		}
	case *tcell.EventMouse:
		if ev.Buttons()&tcell.Button1 != 0 {
			cx, cy := ev.Position()
			cx, cy = cx-m.sx-m.offsetX, cy-m.sy-m.offsetY

			// if last paint time is sufficiently small and the distance is big enough,
			// draw a line
			// otherwise just place a stamp
			dx, dy := cx-b.lastPosition.X, cy-b.lastPosition.Y
			dist := max(max(dx, -dx), max(dy, -dy))
			st := tcell.StyleDefault.Foreground(tcell.ColorRed)
			if ev.When().Sub(b.lastPaint).Seconds() < 0.1 && dist > 1 {
				positions := LinePositions(
					b.lastPosition.X,
					b.lastPosition.Y,
					cx,
					cy,
				)
				for _, p := range positions {
					m.canvas.Set(p.X, p.Y, b.currentIcon, st)
				}
			} else {
				m.canvas.Set(cx, cy, b.currentIcon, st)
			}

			b.lastPaint = ev.When()
			b.lastPosition = Position{X: cx, Y: cy}
		}
	}
}

func (b *BrushTool) Update(m *MainWidget) {
}

func (b *BrushTool) Draw(
	m *MainWidget,
	p Painter,
	x, y, w, h int,
	lag float64,
) {
	SetString(p, x+1, y, "Brush Tool", tcell.StyleDefault)
	p.SetRune(
		m.cursorX+m.sx,
		m.cursorY+m.sy,
		rune(b.currentIcon),
		nil,
		tcell.StyleDefault,
	)
}
