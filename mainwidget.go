package main

import (
	"fmt"
	"os"
	"unicode"

	"github.com/gdamore/tcell/v2"
)

const MAX_BRUSH_RADIUS int = 99

type ColorPickState int

const (
	ColorPickNone ColorPickState = iota
	ColorPickHover
	ColorPickDrag
)

type ColorSelectState int

const (
	ColorSelectNone ColorSelectState = iota
	ColorSelectFg
	ColorSelectBg
)

type LockMask int

const (
	LockMaskAlpha LockMask = 1 << iota
	LockMaskChar
	LockMaskFg
	LockMaskBg
)

var colorMap = map[rune]int{
	'1': 0, '!': 8,
	'2': 1, '@': 9,
	'3': 2, '#': 10,
	'4': 3, '$': 11,
	'5': 4, '%': 12,
	'6': 5, '^': 13,
	'7': 6, '&': 14,
	'8': 7, '*': 15,
	'`': 16,
}

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

	colorPickState   ColorPickState
	colorPickOriginX int
	colorPickOriginY int
	hoverChar        byte
	hoverFg          tcell.Color
	hoverBg          tcell.Color

	colorSelectState ColorSelectState

	hasTool     bool
	currentTool Tool

	statusLine string

	brushCharacter byte
	fgColor        tcell.Color
	bgColor        tcell.Color
	brushRadius    int
	lockMask       LockMask

	clipboard Grid[Cell]

	isStaging     bool
	stagingCanvas *Buffer

	bufferHistory  []*Buffer
	undoHistoryPos int

	isPasting bool
	pasteData []byte
}

var (
	_ Widget = &MainWidget{}
)

func Init(a *App, screen tcell.Screen) *MainWidget {
	w := &MainWidget{
		app:            a,
		canvas:         MakeBuffer(80, 24),
		brushCharacter: '#',
		brushRadius:    1,
	}

	w.ScreenResize(screen.Size())
	w.CenterCanvas()
	w.ClearTool()

	w.cursorX, w.cursorY = w.sw/2, w.sh/2
	return w
}

func (m *MainWidget) HandleEvent(event tcell.Event) {
	if handled := m.HandlePaste(event); handled {
		return
	}

	switch ev := event.(type) {
	case *tcell.EventResize:
		oldsw, oldsh := m.sw, m.sh
		m.ScreenResize(ev.Size())
		m.ScaleOffset(oldsw, oldsh, m.sw, m.sh)
		return
	case *tcell.EventMouse:
		cx, cy := ev.Position()
		cx, cy = cx-m.sx, cy-m.sy
		m.cursorX, m.cursorY = cx, cy
	}

	if handled := m.HandlePan(event); handled {
		return
	}

	if handled := m.HandleColorPick(event); handled {
		return
	}

	if handled := m.HandleColorSelect(event); handled {
		return
	}

	if handled := m.HandleShortcuts(event); handled {
		return
	}

	switch ev := event.(type) {
	case *tcell.EventKey:
		if ev.Key() == tcell.KeyRune {
			r := ev.Rune()
			if r > unicode.MaxASCII {
				r = '?'
			} else if r <= 0x20 {
				r = ' '
			}
			m.brushCharacter = byte(r)
		}
	}

	m.currentTool.HandleEvent(m, event)
}

func (m *MainWidget) StartPaste() {
	m.isPasting = true
	m.pasteData = []byte{}
}

func (m *MainWidget) HandlePaste(event tcell.Event) bool {
	switch ev := event.(type) {
	case *tcell.EventPaste:
		if ev.Start() {
			m.StartPaste()
		} else if m.isPasting && ev.End() {
			// set clipboard and return to stamp tool
			m.SetClipboardFromPasteData()
			m.SetTool(&StampTool{})
			m.isPasting = false
		}
		return true
	case *tcell.EventKey:
		if !m.isPasting {
			return false
		}
		switch ev.Key() {
		case tcell.KeyRune:
			r := ev.Rune()
			if r > unicode.MaxASCII {
				r = '?'
			}
			m.pasteData = append(m.pasteData, byte(r))
		default:
			m.pasteData = append(m.pasteData, '\n')
		}
		return true
	}

	return false
}

func (m *MainWidget) HandlePan(event tcell.Event) bool {
	switch ev := event.(type) {
	case *tcell.EventMouse:
		if ev.Modifiers()&tcell.ModCtrl != 0 &&
			ev.Buttons()&tcell.Button1 != 0 {
			if !m.isPan {
				m.isPan = true
				m.panOriginX, m.panOriginY = m.cursorX, m.cursorY
			}
			return true
		} else if m.isPan {
			m.isPan = false
			m.offsetX += m.cursorX - m.panOriginX
			m.offsetY += m.cursorY - m.panOriginY
			return true
		}
	}
	// If we're panning, then don't run any future event handlers
	return m.isPan
}

func (m *MainWidget) HandleColorPick(event tcell.Event) bool {
	switch ev := event.(type) {
	case *tcell.EventMouse:
		if ev.Modifiers()&tcell.ModAlt != 0 {
			if m.colorPickState != ColorPickDrag {
				m.colorPickState = ColorPickHover
				canvasX, canvasY := m.cursorX-m.offsetX, m.cursorY-m.offsetY

				if m.canvas.Data.InBounds(canvasX, canvasY) {
					cell := m.canvas.Data.MustGet(canvasX, canvasY)
					m.hoverChar = cell.Value
					m.hoverFg, m.hoverBg, _ = cell.Style.Decompose()
				}

				if ev.Buttons()&tcell.Button1 != 0 {
					m.colorPickState = ColorPickDrag
					m.colorPickOriginX = m.cursorX
					m.colorPickOriginY = m.cursorY
					return true
				}
			} else if m.colorPickState == ColorPickDrag {
				if ev.Buttons()&tcell.Button1 == 0 {
					offsetY := m.cursorY - m.colorPickOriginY
					if offsetY < -2 {
						m.fgColor = m.hoverFg
					} else if offsetY > 2 {
						m.bgColor = m.hoverBg
					} else {
						m.brushCharacter = m.hoverChar
					}
					m.colorPickState = ColorPickHover
				}
				return true
			}
			return false
		} else {
			if m.colorPickState == ColorPickDrag {
				offsetY := m.cursorX - m.colorPickOriginY
				if offsetY < -2 {
					m.fgColor = m.hoverFg
				} else if offsetY > 2 {
					m.bgColor = m.hoverBg
				} else {
					m.brushCharacter = m.hoverChar
				}
				m.colorPickState = ColorPickNone
				return true
			}
			m.colorPickState = ColorPickNone
			return false
		}
	}
	return false
}

func (m *MainWidget) HandleColorSelect(event tcell.Event) bool {
	if m.colorSelectState == ColorSelectNone {
		return false
	}
	switch ev := event.(type) {
	case *tcell.EventKey:
		r := ev.Rune()
		if newColor, ok := colorMap[r]; ok {
			if m.colorSelectState == ColorSelectFg {
				m.SetFgColor(newColor)
			} else if m.colorSelectState == ColorSelectBg {
				m.SetBgColor(newColor)
			}
		}
		m.colorSelectState = ColorSelectNone
		return true
	}
	return false
}

func (m *MainWidget) HandleShortcuts(event tcell.Event) bool {
	switch ev := event.(type) {
	case *tcell.EventKey:
		if ev.Modifiers()&tcell.ModAlt != 0 && ev.Key() == tcell.KeyF1 {
			if m.isPan {
				m.offsetX += m.cursorX - m.panOriginX
				m.offsetY += m.cursorY - m.panOriginY
				m.panOriginX = m.cursorX
				m.panOriginY = m.cursorY
			}
			m.CenterCanvas()
			return true
		}

		// FIXME: hack
		if ev.Key() == tcell.KeyEscape {
			m.ClearTool()
			return true
		}

		if ev.Key() == tcell.KeyRune {
			if ev.Modifiers()&tcell.ModAlt != 0 {
				handled := true
				switch ev.Rune() {
				case 'q':
					m.app.WillQuit = true

				case 'h':
					m.SetTool(&HelpTool{})

				case 'p':
					m.SetTool(MakePromptTool(m.Export, "export path..."))

				case 'i':
					m.SetTool(MakePromptTool(m.Import, "import path..."))

				case 's':
					m.SetTool(MakePromptTool(m.Save, "save path..."))

				case 'l':
					m.SetTool(MakePromptTool(m.Load, "load path..."))

				case 'f':
					if m.colorSelectState != ColorSelectFg {
						m.colorSelectState = ColorSelectFg
					} else {
						m.colorSelectState = ColorSelectNone
					}

				case 'g':
					if m.colorSelectState != ColorSelectBg {
						m.colorSelectState = ColorSelectBg
					} else {
						m.colorSelectState = ColorSelectNone
					}

				case 'n':
					m.Stage()
					m.stagingCanvas = MakeBuffer(m.canvas.Data.Width, m.canvas.Data.Height)
					m.Commit()

				case 'r':
					m.SetTool(&LassoTool{})

				case 't':
					m.SetTool(&TranslateTool{})

				case 'e':
					m.SetTool(&LineTool{})

				case 'a':
					m.Stage()
					m.stagingCanvas.Deselect()
					m.Commit()

				case 'c':
					m.SetClipboard()

				case 'x':
					m.SetClipboard()
					m.Stage()
					m.stagingCanvas.ClearSelection()
					m.Commit()

				case 'v':
					m.SetTool(&StampTool{})

				// Increase brush radius
				case '=':
					m.brushRadius = min(MAX_BRUSH_RADIUS, m.brushRadius+1)

				// Decrease brush radius
				case '-':
					m.brushRadius = max(1, m.brushRadius-1)

				// Undo
				case 'z':
					m.undoHistoryPos = min(len(m.bufferHistory), m.undoHistoryPos+1)

				// Redo
				case 'Z':
					m.undoHistoryPos = max(0, m.undoHistoryPos-1)

				case '[':
					t := &ResizeTool{}
					t.SetDimsFromSelection(m.CurrentCanvas())
					m.SetTool(t)

				case '1':
					m.lockMask ^= LockMaskAlpha

				case '2':
					m.lockMask ^= LockMaskChar

				case '3':
					m.lockMask ^= LockMaskFg

				case '4':
					m.lockMask ^= LockMaskBg

				case ',':
					m.Stage()
					m.stagingCanvas.ClearSelection()
					m.Commit()

				case '.':
					m.Stage()
					m.stagingCanvas.FillSelection(Cell{
						Value: m.brushCharacter,
						Style: tcell.StyleDefault.Foreground(m.fgColor).Background(m.bgColor),
					})
					m.Commit()

				default:
					handled = false
				}
				return handled
			}
		}
	}
	return false
}

func (m *MainWidget) Update() {
}

func (m *MainWidget) Draw(p Painter, x, y, w, h int, lag float64) {
	// Draw surrounding box of screen
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

	// Crop to edges of screen
	crop := &CropPainter{
		p:            p,
		offsetBefore: Position{X: r.X, Y: r.Y},
		area:         r,
	}

	// Canvas offset for future drawing operations
	canvasOffX, canvasOffY := m.offsetX, m.offsetY
	if m.isPan {
		canvasOffX += m.cursorX - m.panOriginX
		canvasOffY += m.cursorY - m.panOriginY
	}

	m.DrawCanvas(crop, canvasOffX, canvasOffY)

	// If a tool is active, draw the given tool
	if m.hasTool {
		m.currentTool.Draw(m, p, x, y, w, h, lag)
	} else {
		SetString(p, x+1, y, "No Tool", tcell.StyleDefault)
	}

	m.DrawStatusBar(p, x, y, w, h)

	// color picker
	if m.colorPickState == ColorPickHover {
		cx, cy := m.cursorX+m.sx, m.cursorY+m.sy
		DrawColorPickerState(
			p, cx, cy,
			true, false, false,
			m.hoverChar, m.hoverFg, m.hoverBg,
		)
	} else if m.colorPickState == ColorPickDrag {
		cx, cy := m.cursorX+m.sx, m.cursorY+m.sy
		DrawColorPickDragIndicator(
			p, cx, cy, m.colorPickOriginX+m.sx, m.colorPickOriginY+m.sy,
		)
	} else if m.IsPaintTool() {
		cx, cy := m.cursorX+m.sx, m.cursorY+m.sy
		FillRegion(
			p,
			cx-m.brushRadius/2, cy-m.brushRadius/2,
			m.brushRadius, m.brushRadius,
			rune(m.brushCharacter),
			tcell.StyleDefault.Foreground(m.fgColor).Background(m.bgColor),
		)
		p.SetByte(cx, cy, m.brushCharacter, tcell.StyleDefault.Foreground(m.fgColor).Background(m.bgColor))
	}

	// color selector
	if m.colorSelectState != ColorSelectNone {
		DrawColorSelector(p, x, y+1, m.colorSelectState)
	}
}

func (m *MainWidget) DrawStatusBar(p Painter, x, y, w, h int) {
	// Draw the statusline
	SetCenteredString(p, x+w/2, y, m.statusLine, tcell.StyleDefault)

	// color/char indicators
	SetString(p, x+w-27, y, fmt.Sprintf("radius: %d", m.brushRadius), tcell.StyleDefault)
	SetString(p, x+w-17, y, "char: ", tcell.StyleDefault)
	p.SetByte(x+w-12, y, m.brushCharacter, tcell.StyleDefault)
	SetString(p, x+w-10, y, "fg: ", tcell.StyleDefault)
	DrawColorSymbolFG(p, x+w-7, y, m.fgColor)
	SetString(p, x+w-5, y, "bg: ", tcell.StyleDefault)
	DrawColorSymbolBG(p, x+w-2, y, m.bgColor)

	// Lock mask
	SetString(p, x+w-38, y, fmt.Sprintf("lock: ____"), tcell.StyleDefault)
	if m.lockMask&LockMaskAlpha != 0 {
		p.SetByte(x+w-32, y, 'a', tcell.StyleDefault)
	}
	if m.lockMask&LockMaskChar != 0 {
		p.SetByte(x+w-31, y, 'c', tcell.StyleDefault)
	}
	if m.lockMask&LockMaskFg != 0 {
		p.SetByte(x+w-30, y, 'f', tcell.StyleDefault)
	}
	if m.lockMask&LockMaskBg != 0 {
		p.SetByte(x+w-29, y, 'g', tcell.StyleDefault)
	}
}

func (m *MainWidget) DrawCanvas(p Painter, offX, offY int) {
	// Rendering the canvas
	curCanvas := m.canvas
	if m.isStaging {
		curCanvas = m.stagingCanvas
	} else if m.undoHistoryPos > 0 {
		curCanvas = m.bufferHistory[len(m.bufferHistory)-m.undoHistoryPos]
	}
	curCanvas.RenderWith(p, offX, offY, true)

	BorderBox(p, Area{
		X:      offX - 1,
		Y:      offY - 1,
		Width:  m.canvas.Data.Width + 2,
		Height: m.canvas.Data.Height + 2,
	}, tcell.StyleDefault)

	// selection mask
	if curCanvas.activeSelection {
		for y := range curCanvas.Data.Height {
			for x := range curCanvas.Data.Width {
				if curCanvas.SelectionMask.MustGet(x, y) {
					xx, yy := x+offX, y+offY
					_, s := p.GetContent(xx, yy)
					p.SetStyle(xx, yy, s.Reverse(true))
				}
			}
		}
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

func (m *MainWidget) ResizeCanvas(newRect Area) {
	m.Stage()
	m.stagingCanvas = MakeBuffer(newRect.Width, newRect.Height)
	for y := range m.canvas.Data.Height {
		for x := range m.canvas.Data.Width {
			nx, ny := x-newRect.X, y-newRect.Y
			m.stagingCanvas.Data.Set(nx, ny, m.canvas.Data.MustGet(x, y))
			m.stagingCanvas.SelectionMask.Set(nx, ny, m.canvas.SelectionMask.MustGet(x, y))
		}
	}
	m.Commit()
	m.offsetX += newRect.X
	m.offsetY += newRect.Y
	m.ClearTool()
}

func (m *MainWidget) CenterCanvas() {
	cw, ch := m.canvas.Data.Width, m.canvas.Data.Height
	m.offsetX = (m.sw - cw) / 2
	m.offsetY = (m.sh - ch) / 2
}

func (m *MainWidget) SetTool(tool Tool) {
	m.Rollback()
	m.hasTool = true
	m.currentTool = tool
	m.statusLine = ""
}

func (m *MainWidget) ClearTool() {
	m.Rollback()
	m.hasTool = true
	m.currentTool = &BrushTool{}
	m.statusLine = ""
}

func (m *MainWidget) SetFgColor(fg int) {
	if fg == 16 {
		m.fgColor = tcell.ColorDefault
	} else {
		m.fgColor = tcell.ColorValid + tcell.Color(fg)
	}
}

func (m *MainWidget) SetBgColor(bg int) {
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
	m.Stage()
	var msg string
	defer func() {
		m.ClearTool()
		m.statusLine = msg
		return
	}()

	f, err := os.Open(s)
	if err != nil {
		m.Rollback()
		msg = err.Error()
		return
	}
	if err := m.stagingCanvas.Import(f); err != nil {
		m.Rollback()
		msg = err.Error()
		return
	}

	m.Commit()

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
	m.Stage()
	var msg string
	defer func() {
		m.ClearTool()
		m.statusLine = msg
		return
	}()

	f, err := os.Open(s)
	if err != nil {
		msg = err.Error()
		m.Rollback()
		return
	}
	if err := m.stagingCanvas.Load(f); err != nil {
		msg = err.Error()
		m.Rollback()
		return
	}

	m.Commit()

	msg = fmt.Sprintf("Successfully loaded %s", s)
}

func (m *MainWidget) ReplaceSelectionMask(topLeft Position, mask Grid[bool]) {
	m.canvas.SetSelection(mask, topLeft)
}

func (m *MainWidget) SetClipboard() {
	m.clipboard = m.canvas.CopySelection()
}

func (m *MainWidget) SetClipboardFromPasteData() {
	var width, height int
	var w = 0
	for _, c := range m.pasteData {
		if c == '\n' {
			height++
			w = 0
		} else {
			w++
			width = max(width, w)
		}
	}
	clip := MakeGrid(width+1, height+1, Cell{Value: ' '})

	var x, y int
	for _, c := range m.pasteData {
		if c == '\n' {
			x, y = 0, y+1
		} else {
			clip.Set(x, y, Cell{
				Value: c,
				Style: tcell.StyleDefault.Foreground(m.fgColor).Background(m.bgColor),
			})
			x++
		}
	}

	m.clipboard = clip
}

func (m *MainWidget) Stage() {
	if m.isStaging {
		return
	}

	curCanvas := m.canvas
	if m.undoHistoryPos > 0 {
		curCanvas = m.bufferHistory[len(m.bufferHistory)-m.undoHistoryPos]
	}
	m.isStaging = true
	m.stagingCanvas = curCanvas.Clone()
}

func (m *MainWidget) Commit() {
	if !m.isStaging {
		return
	}

	if m.undoHistoryPos > 0 {
		m.bufferHistory = m.bufferHistory[:len(m.bufferHistory)-(m.undoHistoryPos-1)]
		m.undoHistoryPos = 0
	} else {
		m.bufferHistory = append(m.bufferHistory, m.canvas)
	}
	m.isStaging = false
	m.canvas = m.stagingCanvas
	m.stagingCanvas = nil
}

func (m *MainWidget) Rollback() {
	if !m.isStaging {
		return
	}
	m.isStaging = false
	m.stagingCanvas = nil
}

func (m *MainWidget) CurrentCanvas() *Buffer {
	if m.isStaging {
		return m.stagingCanvas
	} else if m.undoHistoryPos > 0 {
		return m.bufferHistory[len(m.bufferHistory)-m.undoHistoryPos]
	} else {
		return m.canvas
	}
}

func (m *MainWidget) IsPaintTool() bool {
	if _, ok := m.currentTool.(*BrushTool); ok {
		return true
	}
	if _, ok := m.currentTool.(*LineTool); ok {
		return true
	}
	return false
}
