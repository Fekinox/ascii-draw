package main

import (
	"fmt"
	"os"

	"github.com/gdamore/tcell/v2"
)

type MainWidget struct {
	app *App

	sx int
	sy int
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

	hasTool     bool
	currentTool Tool

	statusLine string

	brushCharacter byte
	fgColor        tcell.Color
	bgColor        tcell.Color
}

var (
	_ Widget = &MainWidget{}
)

func Init(a *App, screen tcell.Screen) *MainWidget {
	w := &MainWidget{
		app:    a,
		canvas: MakeBuffer(10, 6),
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
		cx, cy = cx-m.sx, cy-m.sy
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
		if ev.Modifiers()&tcell.ModAlt != 0 && ev.Key() == tcell.KeyF1 {
			if m.isPan {
				m.offsetX += m.cursorX - m.panOriginX
				m.offsetY += m.cursorY - m.panOriginY
				m.panOriginX = m.cursorX
				m.panOriginY = m.cursorY
			}
			m.CenterCanvas()
			return
		}

		if ev.Key() == tcell.KeyEscape {
			if m.hasTool {
				m.hasTool = false
				m.currentTool = nil
			} else {
				m.app.WillQuit = true
			}
			return
		}

		if ev.Modifiers()&tcell.ModAlt != 0 && ev.Rune() == 'b' {
			m.SetTool(&BrushTool{currentIcon: '#'})
			return
		}

		if ev.Modifiers()&tcell.ModAlt != 0 && ev.Rune() == 'x' {
			m.SetTool(MakePromptTool(m.Export, "export path..."))
			return
		}

		if ev.Modifiers()&tcell.ModAlt != 0 && ev.Rune() == 'i' {
			m.SetTool(MakePromptTool(m.Import, "import path..."))
			return
		}

		if ev.Modifiers()&tcell.ModAlt != 0 && ev.Rune() == 's' {
			m.SetTool(MakePromptTool(m.Save, "save path..."))
			return
		}

		if ev.Modifiers()&tcell.ModAlt != 0 && ev.Rune() == 'l' {
			m.SetTool(MakePromptTool(m.Load, "load path..."))
			return
		}

		if ev.Modifiers()&tcell.ModAlt != 0 && ev.Rune() == 'c' {
			m.SetTool(MakeColorSelectorTool())
			return
		}

		if ev.Modifiers()&tcell.ModAlt != 0 && ev.Rune() == 'p' {
			m.SetTool(MakeColorPickerTool())
			return
		}
	}

	if !m.isPan && m.hasTool {
		m.currentTool.HandleEvent(m, event)
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
		Width:  m.canvas.Data.Width + 2,
		Height: m.canvas.Data.Height + 2,
	}, tcell.StyleDefault)

	if m.hasTool {
		m.currentTool.Draw(m, p, x, y, w, h, lag)
	} else {
		SetString(p, x+1, y, "No Tool", tcell.StyleDefault)
	}

	SetCenteredString(p, x+w/2, y, m.statusLine, tcell.StyleDefault)

	// color/char indicators
	SetString(p, x+w-17, y, "char: ", tcell.StyleDefault)
	p.SetByte(x+w-12, y, m.brushCharacter, tcell.StyleDefault)
	SetString(p, x+w-10, y, "fg: ", tcell.StyleDefault)
	if m.fgColor == 0 {
		p.SetByte(x+w-7, y, '_', tcell.StyleDefault)
	} else {
		var c byte = 'b'
		if m.fgColor <= tcell.ColorGray {
			c = 'n'
		}
		p.SetByte(x+w-7, y, c, tcell.StyleDefault.Background(m.fgColor))
	}
	SetString(p, x+w-5, y, "bg: ", tcell.StyleDefault)
	if m.bgColor == 0 {
		p.SetByte(x+w-2, y, '_', tcell.StyleDefault)
	} else {
		var c byte = 'b'
		if m.bgColor <= tcell.ColorGray {
			c = 'n'
		}
		p.SetByte(x+w-2, y, c, tcell.StyleDefault.Background(m.bgColor))
	}
}

func (m *MainWidget) ScreenResize(sw, sh int) {
	m.sw, m.sh = sw-2, sh-2
	m.sx, m.sy = 1, 1
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

func (m *MainWidget) SetTool(tool Tool) {
	m.hasTool = true
	m.currentTool = tool
	m.statusLine = ""
}

func (m *MainWidget) ClearTool() {
	m.hasTool = false
	m.currentTool = nil
	m.statusLine = ""
}

func (m *MainWidget) SetColor(fg, bg int) {
	if fg == 16 {
		m.fgColor = tcell.ColorDefault
	} else {
		m.fgColor = tcell.ColorValid + tcell.Color(fg)
	}

	if bg == 16 {
		m.bgColor = tcell.ColorDefault
	} else {
		m.bgColor = tcell.ColorValid + tcell.Color(bg)
	}
}

func (m *MainWidget) Export(s string) {
	var msg string
	defer func() {
		m.ClearTool()
		m.statusLine = msg
		return
	}()

	f, err := os.Create(s)
	if err != nil {
		msg = err.Error()
		return
	}
	if err := m.canvas.Export(f); err != nil {
		msg = err.Error()
		return
	}

	msg = fmt.Sprintf("Successfully exported to plaintext file %s", s)
}

func (m *MainWidget) Import(s string) {
	var msg string
	defer func() {
		m.ClearTool()
		m.statusLine = msg
		return
	}()

	f, err := os.Open(s)
	if err != nil {
		msg = err.Error()
		return
	}
	if err := m.canvas.Import(f); err != nil {
		msg = err.Error()
		return
	}

	msg = fmt.Sprintf("Successfully imported plaintext file %s", s)
}

func (m *MainWidget) Save(s string) {
	var msg string
	defer func() {
		m.ClearTool()
		m.statusLine = msg
		return
	}()

	f, err := os.Create(s)
	if err != nil {
		msg = err.Error()
		return
	}
	if err := m.canvas.Save(f); err != nil {
		msg = err.Error()
		return
	}

	msg = fmt.Sprintf("Successfully saved %s", s)
}

func (m *MainWidget) Load(s string) {
	var msg string
	defer func() {
		m.ClearTool()
		m.statusLine = msg
		return
	}()

	f, err := os.Open(s)
	if err != nil {
		msg = err.Error()
		return
	}
	if err := m.canvas.Load(f); err != nil {
		msg = err.Error()
		return
	}

	msg = fmt.Sprintf("Successfully loaded %s", s)
}
