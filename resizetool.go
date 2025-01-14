package main

import (
	"math"

	"github.com/gdamore/tcell/v2"
)

type ResizeToolState int

const (
	ResizeToolNone ResizeToolState = iota
	ResizeToolMoving
	ResizeToolCreating
)

type ResizeTool struct {
	state ResizeToolState
	dims  Area

	origX int
	origY int
}

func lineDistance(ax, ay, bx, by, x, y int) int {
	dlx, dly := bx-ax, by-ay
	dpx, dpy := x-ax, y-ay

	param := float64(dlx*dpx+dly*dpy) / float64(dlx*dlx+dly*dly)
	projX, projY := float64(dlx)*param, float64(dly)*param
	perpX, perpY := float64(dpx)-projX, float64(dpy)-projY

	if param < 0 || param > 1 {
		return int(math.Sqrt(float64(min(
			(x-ax)*(x-ax)+(y-ay)*(y-ay),
			(x-bx)*(x-bx)+(y-by)*(y-by)))))
	} else {
		return int(math.Sqrt(perpX*perpX + perpY*perpY))
	}
}

func (r *ResizeTool) SetDimsFromSelection(b *Buffer) {
	if !b.activeSelection {
		r.dims.X, r.dims.Y = 0, 0
		r.dims.Width, r.dims.Height = b.Data.Width, b.Data.Height
		return
	}

	minX, maxX := b.Data.Width-1, 0
	minY, maxY := b.Data.Height-1, 0

	for y := range b.Data.Height {
		for x := range b.Data.Width {
			if b.SelectionMask.MustGet(x, y) {
				minX, maxX = min(minX, x), max(maxX, x)
				minY, maxY = min(minY, y), max(maxY, y)
			}
		}
	}

	r.dims.X, r.dims.Y = minX, minY
	r.dims.Width, r.dims.Height = maxX-minX+1, maxY-minY+1
}

func (r *ResizeTool) HandleEvent(m *MainWidget, event tcell.Event) {
	switch ev := event.(type) {
	case *tcell.EventMouse:
		cx, cy := m.cursorX-m.offsetX, m.cursorY-m.offsetY
		_, _ = cx, cy
		if ev.Buttons()&tcell.Button1 != 0 {
			if r.state == ResizeToolNone && r.dims.Contains(cx, cy) {
				r.state = ResizeToolMoving
				r.origX, r.origY = cx, cy
			} else if r.state == ResizeToolNone {
				r.state = ResizeToolCreating
				r.origX, r.origY = cx, cy
			}
		} else if r.state == ResizeToolMoving {
			r.state = ResizeToolNone
			r.dims.X += cx - r.origX
			r.dims.Y += cy - r.origY
		} else if r.state == ResizeToolCreating {
			r.state = ResizeToolNone

			minX, maxX := min(cx, r.origX), max(cx, r.origX)
			minY, maxY := min(cy, r.origY), max(cy, r.origY)
			r.dims.X, r.dims.Y = minX, minY
			r.dims.Width, r.dims.Height = maxX-minX+1, maxY-minY+1
		}
	}
}

func (r *ResizeTool) Draw(m *MainWidget, p Painter, x, y, w, h int, lag float64) {
	crop := &CropPainter{
		p: p,
		area: Area{
			X:      m.sx,
			Y:      m.sy,
			Width:  m.sw,
			Height: m.sh,
		},
	}

	cx, cy := m.cursorX-m.offsetX, m.cursorY-m.offsetY
	ox, oy := m.offsetX, m.offsetY
	if m.isPan {
		ox += m.cursorX - m.panOriginX
		oy += m.cursorY - m.panOriginY
	}
	if r.state == ResizeToolMoving {
		ox += cx - r.origX
		oy += cy - r.origY
	}

	d := r.dims
	if r.state == ResizeToolCreating {
		minY, maxY := min(cy, r.origY), max(cy, r.origY)
		minX, maxX := min(cx, r.origX), max(cx, r.origX)
		d.X, d.Y = minX, minY
		d.Width, d.Height = maxX-minX+1, maxY-minY+1
	}
	a := Area{
		X:      x + m.sx + ox + d.X,
		Y:      y + m.sy + oy + d.Y,
		Width:  d.Width,
		Height: d.Height,
	}
	BorderBox(crop, a, tcell.StyleDefault.Foreground(tcell.ColorRed))
}
