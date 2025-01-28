package main

import (
	"fmt"

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
	ResizeToolTop ResizeToolEdge = 1 << iota
	ResizeToolBottom
	ResizeToolLeft
	ResizeToolRight
)

type ResizeToolHandle struct {
	pos   Position
	edges ResizeToolEdge
}

type ResizeTool struct {
	state       ResizeToolState
	dims        Area
	stagingDims Area
	handles     [8]ResizeToolHandle

	origX      int
	origY      int
	expandEdge ResizeToolEdge
}

func (r *ResizeTool) InitHandles() {
	// top left
	r.handles[0] = ResizeToolHandle{
		edges: ResizeToolTop | ResizeToolLeft,
	}
	// top
	r.handles[1] = ResizeToolHandle{
		edges: ResizeToolTop,
	}
	// top right
	r.handles[2] = ResizeToolHandle{
		edges: ResizeToolTop | ResizeToolRight,
	}
	// right
	r.handles[3] = ResizeToolHandle{
		edges: ResizeToolRight,
	}
	// bottom right
	r.handles[4] = ResizeToolHandle{
		edges: ResizeToolBottom | ResizeToolRight,
	}
	// bottom
	r.handles[5] = ResizeToolHandle{
		edges: ResizeToolBottom,
	}
	// bottom left
	r.handles[6] = ResizeToolHandle{
		edges: ResizeToolBottom | ResizeToolLeft,
	}
	// left
	r.handles[7] = ResizeToolHandle{
		edges: ResizeToolLeft,
	}
	r.RepositionHandles(r.dims)
}

func (r *ResizeTool) RepositionHandles(d Area) {
	// top left
	r.handles[0].pos = Position{
		X: d.X - 2,
		Y: d.Y - 2,
	}
	// top
	r.handles[1].pos = Position{
		X: d.X + d.Width/2 + 1,
		Y: d.Y - 2,
	}
	// top right
	r.handles[2].pos = Position{
		X: d.X + d.Width + 3,
		Y: d.Y - 2,
	}
	// right
	r.handles[3].pos = Position{
		X: d.X + d.Width + 3,
		Y: d.Y + d.Height/2 + 1,
	}
	// bottom right
	r.handles[4].pos = Position{
		X: d.X + d.Width + 3,
		Y: d.Y + d.Height + 3,
	}
	// bottom
	r.handles[5].pos = Position{
		X: d.X + d.Width/2 + 1,
		Y: d.Y + d.Height + 3,
	}
	// bottom left
	r.handles[6].pos = Position{
		X: d.X - 2,
		Y: d.Y + d.Height + 3,
	}
	// left
	r.handles[7].pos = Position{
		X: d.X - 2,
		Y: d.Y + d.Height/2 + 1,
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
			if r.state == ResizeToolNone {
				r.stagingDims = r.dims
				r.origX, r.origY = cx, cy

				minDist := SquaredDistance(cx, cy, r.handles[0].pos.X, r.handles[0].pos.Y)
				var argmin int
				for j, h := range r.handles {
					d := SquaredDistance(cx, cy, h.pos.X, h.pos.Y)
					if d < minDist {
						minDist = d
						argmin = j
					}
				}

				if minDist <= 4 {
					r.state = ResizeToolExpanding
					r.expandEdge = r.handles[argmin].edges
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
				r.RepositionHandles(r.stagingDims)
			} else if r.state == ResizeToolCreating {

				minX, maxX := min(cx, r.origX), max(cx, r.origX)
				minY, maxY := min(cy, r.origY), max(cy, r.origY)
				r.stagingDims.X, r.stagingDims.Y = minX, minY
				r.stagingDims.Width, r.stagingDims.Height = maxX-minX+1, maxY-minY+1
				r.RepositionHandles(r.stagingDims)
			} else if r.state == ResizeToolExpanding {
				if r.expandEdge&ResizeToolLeft != 0 {
					if r.dims.Width-dx >= 1 {
						r.stagingDims.X = r.dims.X + dx
						r.stagingDims.Width = r.dims.Width - dx
					} else {
						r.stagingDims.X = r.dims.X + r.dims.Width
						r.stagingDims.Width = dx - r.dims.Width
					}
				}

				if r.expandEdge&ResizeToolRight != 0 {
					if r.dims.Width+dx >= 1 {
						r.stagingDims.Width = r.dims.Width + dx
					} else {
						r.stagingDims.X = r.dims.X + r.dims.Width + dx
						r.stagingDims.Width = -r.dims.Width - dx
					}
				}
				if r.expandEdge&ResizeToolTop != 0 {
					if r.dims.Height-dy >= 1 {
						r.stagingDims.Y = r.dims.Y + dy
						r.stagingDims.Height = r.dims.Height - dy
					} else {
						r.stagingDims.Y = r.dims.Y + r.dims.Height
						r.stagingDims.Height = dy - r.dims.Height
					}
				}
				if r.expandEdge&ResizeToolBottom != 0 {
					if r.dims.Height+dy >= 1 {
						r.stagingDims.Height = r.dims.Height + dy
					} else {
						r.stagingDims.Y = r.dims.Y + r.dims.Height + dy
						r.stagingDims.Height = -r.dims.Height - dy
					}
				}
				r.RepositionHandles(r.stagingDims)
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
	SetString(p, x+m.sx, y+m.sy-1, "Resize Tool", tcell.StyleDefault)
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

	for _, h := range r.handles {
		crop.SetByte(h.pos.X+ox, h.pos.Y+oy, '@', tcell.StyleDefault)
	}

	dimsString := fmt.Sprintf("%d x %d", d.Width, d.Height)
	dimsStringX, dimsStringY := x+m.sx+ox+d.X-1, y+m.sy+oy+d.Y-2
	dimsStringX = max(m.sx, min(m.sx+m.sw-len(dimsString), dimsStringX))
	dimsStringY = max(m.sy, min(m.sy+m.sh-1, dimsStringY))
	SetString(crop, dimsStringX, dimsStringY, dimsString, tcell.StyleDefault)

}
