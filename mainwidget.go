package main

import (
	"github.com/gdamore/tcell/v2"
)

type MainWidget struct {
	sw int
	sh int

	cursorX int
	cursorY int

	// Position of top-left corner of buffer
	offsetX int
	offsetY int

	canvas *Buffer

	isPan      bool
	panOriginX int
	panOriginY int
}

var (
	_ Widget = &MainWidget{}
)

func Init(screen tcell.Screen) *MainWidget {
	w := &MainWidget{
		canvas: MakeBuffer(100, 60),
	}

	w.ScreenResize(screen.Size())
	w.CenterCanvas()
	return w
}

func (m *MainWidget) HandleAction(action Action) {
	switch action {
	case MoveUp:
		m.cursorY--
	case MoveDown:
		m.cursorY++
	case MoveLeft:
		m.cursorX--
	case MoveRight:
		m.cursorX++
	}
}

func (m *MainWidget) HandleEvent(event tcell.Event) {
	switch ev := event.(type) {
	case *tcell.EventResize:
		oldsw, oldsh := m.sw, m.sh
		m.ScreenResize(ev.Size())
		m.ScaleOffset(oldsw, oldsh, m.sw, m.sh)
	case *tcell.EventMouse:
		cx, cy := ev.Position()
		m.cursorX, m.cursorY = cx, cy

		if ev.Modifiers()&tcell.ModAlt != 0 &&
			ev.Buttons()&tcell.Button1 != 0 {
			if !m.isPan {
				m.isPan = true
				m.panOriginX, m.panOriginY = cx, cy
			}
		} else if m.isPan {
			m.isPan = false
			m.offsetX += m.cursorX - m.panOriginX
			m.offsetY += m.cursorY - m.panOriginY
		}
	case *tcell.EventKey:
		if m.isPan {
			m.offsetX += m.cursorX - m.panOriginX
			m.offsetY += m.cursorY - m.panOriginY
			m.panOriginX = m.cursorX
			m.panOriginY = m.cursorY
		}
		if ev.Modifiers()&tcell.ModAlt != 0 && ev.Key() == tcell.KeyF1 {
			m.CenterCanvas()
		}
	}
}

func (m *MainWidget) Update() {
}

func (m *MainWidget) Draw(p Painter, x, y, w, h int, lag float64) {
	r := Area{
		X:      x + 1,
		Y:      y + 1,
		Width:  w - 2,
		Height: h - 2,
	}

	BorderBox(p, Area{
		X:      r.X - 1,
		Y:      r.Y - 1,
		Width:  r.Width + 2,
		Height: r.Height + 2,
	}, tcell.StyleDefault)

	crop := &CropPainter{
		p:            p,
		offsetBefore: Position{X: r.X, Y: r.Y},
		area:         r,
	}

	canvasOffX, canvasOffY := m.offsetX, m.offsetY
	if m.isPan {
		canvasOffX += m.cursorX - m.panOriginX
		canvasOffY += m.cursorY - m.panOriginY
	}

	m.canvas.RenderWith(crop, canvasOffX, canvasOffY, true)
	BorderBox(crop, Area{
		X:      canvasOffX - 1,
		Y:      canvasOffY - 1,
		Width:  m.canvas.Data.Width + 1,
		Height: m.canvas.Data.Height + 1,
	}, tcell.StyleDefault)

	p.SetRune(m.cursorX, m.cursorY, '#', nil, tcell.StyleDefault)
}

func (m *MainWidget) ScreenResize(sw, sh int) {
	m.sw, m.sh = sw-2, sh-2
}

func (m *MainWidget) ScaleOffset(oldsw, oldsh, newsw, newsh int) {
	ocx, ocy := m.offsetX-oldsw/2, m.offsetY-oldsh/2
	sfx, sfy := float64(newsw)/float64(oldsw), float64(newsh)/float64(oldsh)
	m.offsetX = int(float64(ocx)*sfx) + newsw/2
	m.offsetY = int(float64(ocy)*sfy) + newsh/2
}

func (m *MainWidget) CenterCanvas() {
	cw, ch := m.canvas.Data.Width, m.canvas.Data.Height
	m.offsetX = (m.sw - cw) / 2
	m.offsetY = (m.sh - ch) / 2
}
