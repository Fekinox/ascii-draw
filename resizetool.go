package main

import (
	"math"
	"slices"

	"github.com/gdamore/tcell/v2"
)

type ResizeToolState int
type ResizeToolEdge int

const (
	ResizeToolNone ResizeToolState = iota
	ResizeToolMoving
	ResizeToolCreating
	ResizeToolExpanding
)

const (
	ResizeToolTop ResizeToolEdge = iota
	ResizeToolBottom
	ResizeToolLeft
	ResizeToolRight
)

type EdgeDistance struct {
	Edge     ResizeToolEdge
	Distance int
}

type ResizeTool struct {
	state       ResizeToolState
	dims        Area
	stagingDims Area

	origX      int
	origY      int
	expandEdge ResizeToolEdge
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

func edgeDistances(rect Area, x, y int) [4]EdgeDistance {
	var res [4]EdgeDistance
	res[0] = EdgeDistance{
		Edge: ResizeToolLeft,
		Distance: lineDistance(
			rect.X, rect.Y,
			rect.X, rect.Y+rect.Height-1,
			x, y,
		),
	}
	res[1] = EdgeDistance{
		Edge: ResizeToolTop,
		Distance: lineDistance(
			rect.X, rect.Y,
			rect.X+rect.Width-1, rect.Y,
			x, y,
		),
	}
	res[2] = EdgeDistance{
		Edge: ResizeToolBottom,
		Distance: lineDistance(
			rect.X, rect.Y+rect.Height-1,
			rect.X+rect.Width-1, rect.Y+rect.Height-1,
			x, y,
		),
	}
	res[3] = EdgeDistance{
		Edge: ResizeToolRight,
		Distance: lineDistance(
			rect.X+rect.Width-1, rect.Y,
			rect.X+rect.Width-1, rect.Y+rect.Height-1,
			x, y,
		),
	}

	slices.SortFunc(res[:], func(e1, e2 EdgeDistance) int {
		return e1.Distance - e2.Distance
	})

	return res
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
			if r.state == ResizeToolNone {
				r.stagingDims = r.dims
				r.origX, r.origY = cx, cy

				distances := edgeDistances(r.dims, cx, cy)

				for _, d := range distances {
					if d.Distance > 3 {
						break
					}
					if d.Distance < 3 {
						r.state = ResizeToolExpanding
						r.expandEdge = d.Edge
						break
					}
				}
			}
			if r.state == ResizeToolNone {
				if r.dims.Contains(cx, cy) {
					r.state = ResizeToolMoving
				} else {
					r.state = ResizeToolCreating
				}
			}

			dx, dy := cx-r.origX, cy-r.origY
			if r.state == ResizeToolMoving {
				r.stagingDims.X = r.dims.X + dx
				r.stagingDims.Y = r.dims.Y + dy
			} else if r.state == ResizeToolCreating {

				minX, maxX := min(cx, r.origX), max(cx, r.origX)
				minY, maxY := min(cy, r.origY), max(cy, r.origY)
				r.stagingDims.X, r.stagingDims.Y = minX, minY
				r.stagingDims.Width, r.stagingDims.Height = maxX-minX+1, maxY-minY+1
			} else if r.state == ResizeToolExpanding {
				switch r.expandEdge {
				case ResizeToolLeft:
					r.stagingDims.X = r.dims.X + dx
					r.stagingDims.Width = r.dims.Width - dx
				case ResizeToolRight:
					r.stagingDims.Width = r.dims.Width + dx
				case ResizeToolTop:
					r.stagingDims.Y = r.dims.Y + dy
					r.stagingDims.Height = r.dims.Height - dy
				case ResizeToolBottom:
					r.stagingDims.Height = r.dims.Height + dy
				}
			}
		} else if r.state != ResizeToolNone {
			r.dims = r.stagingDims
			r.state = ResizeToolNone
		}
	case *tcell.EventKey:
		if r.state != ResizeToolNone {
			r.dims = r.stagingDims
			r.state = ResizeToolNone
		}
		if ev.Key() == tcell.KeyEnter {
			m.ResizeCanvas(r.dims)
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

	ox, oy := m.offsetX, m.offsetY
	if m.isPan {
		ox += m.cursorX - m.panOriginX
		oy += m.cursorY - m.panOriginY
	}

	d := r.dims
	if r.state != ResizeToolNone {
		d = r.stagingDims
	}
	a := Area{
		X:      x + m.sx + ox + d.X - 1,
		Y:      y + m.sy + oy + d.Y - 1,
		Width:  d.Width + 2,
		Height: d.Height + 2,
	}
	BorderBox(crop, a, tcell.StyleDefault.Foreground(tcell.ColorRed))
}
