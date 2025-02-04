package main

import (
	"fmt"
	"math"
	"time"
	"unicode"

	action "github.com/Fekinox/ascii-draw/internal"
	"github.com/gdamore/tcell/v2"
)

const MAX_BRUSH_RADIUS int = 99
const INIT_WIDTH = 80
const INIT_HEIGHT = 24

const EVENT_MONITORING = false

type ColorPickState int

const (
	// Not currently color picking.
	ColorPickNone ColorPickState = iota
	// User is hovering over a cell in the grid to sample its color.
	ColorPickHover
	// User is clicking and dragging on a cell.
	ColorPickDrag
)

type ColorSelectState int

const (
	// Color selector is inactive.
	ColorSelectNone ColorSelectState = iota
	// Color selector for the foreground is active.
	ColorSelectFg
	// Color selector for the background is active.
	ColorSelectBg
)

type LockMask int

const (
	// If set, makes painting operations ignore empty cells. Empty cells are cells where the
	// character is the space character (` `) and where the background color is the terminal
	// default.
	LockMaskAlpha LockMask = 1 << iota
	// If set, painting operations do not modify the character of a cell.
	LockMaskChar
	// If set, painting operations do not modify the foreground color of a cell.
	LockMaskFg
	// If set, painting operations do not modify the background color of a cell.
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

type Editor struct {
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

	hasModalTool     bool
	currentModalTool Tool

	brushCharacter byte
	fgColor        tcell.Color
	bgColor        tcell.Color
	brushRadius    int
	lockMask       LockMask

	clipboard Grid[Cell]

	isStaging     bool
	stagingCanvas *Buffer

	undoHistory    []*Buffer
	undoHistoryPos int
	historyChanged bool

	isPasting        bool
	pendingPasteData []byte

	savedFile string
	// Position of the currently saved editor state in the undo history
	savedUndoIndex int

	appStartTime time.Time

	notification NotificationHandler

	keymap map[KeyEvent]action.Action
}

var (
	_ Widget = &Editor{}
)

func Init(a *App, screen tcell.Screen) *Editor {
	w := &Editor{
		app:            a,
		canvas:         MakeBuffer(INIT_WIDTH, INIT_HEIGHT),
		brushCharacter: '#',
		brushRadius:    1,
		appStartTime:   time.Now(),
		notification:   &NotificationWidget{},
		keymap:         defaultKeymap(),
	}

	w.ScreenResize(screen.Size())
	w.CenterCanvas()
	w.ClearTool()

	w.cursorX, w.cursorY = w.sw/2, w.sh/2
	a.Logger.Println("Successfully initialized program")
	return w
}

func defaultKeymap() map[KeyEvent]action.Action {
	return map[KeyEvent]action.Action{
		{Key: tcell.KeyCtrlQ}: action.Quit,
		{Key: tcell.KeyCtrlQ,
			Modifiers: tcell.ModAlt}: action.ForceQuit,
		{Key: tcell.KeyCtrlH}: action.Help,

		{Key: tcell.KeyCtrlS}: action.Save,
		{Key: tcell.KeyCtrlO}: action.Load,
		{Key: tcell.KeyCtrlP}: action.Export,
		{Key: tcell.KeyCtrlI}: action.Import,
		{Key: tcell.KeyCtrlN}: action.NewCanvas,

		{Key: tcell.KeyCtrlF}: action.FgColorSelector,
		{Key: tcell.KeyCtrlG}: action.BgColorSelector,

		{Key: tcell.KeyCtrlR}: action.Lasso,
		{Key: tcell.KeyCtrlT}: action.Translate,
		{Key: tcell.KeyCtrlE}: action.Line,

		{Key: tcell.KeyCtrlA}:        action.Deselect,
		{Key: tcell.KeyCtrlC}:        action.Copy,
		{Key: tcell.KeyCtrlX}:        action.Cut,
		{Key: tcell.KeyCtrlV}:        action.Paste,
		RuneEvent(',', tcell.ModAlt): action.ClearSelection,
		RuneEvent('.', tcell.ModAlt): action.FillSelection,

		{Key: tcell.KeyCtrlZ}: action.Undo,
		{Key: tcell.KeyCtrlY}: action.Redo,
		{Key: tcell.KeyF1,
			Modifiers: tcell.ModAlt}: action.CenterCanvas,

		RuneEvent('=', tcell.ModAlt): action.IncreaseBrushRadius,
		RuneEvent('-', tcell.ModAlt): action.DecreaseBrushRadius,
		RuneEvent('1', tcell.ModAlt): action.AlphaLock,
		RuneEvent('2', tcell.ModAlt): action.CharLock,
		RuneEvent('3', tcell.ModAlt): action.FgLock,
		RuneEvent('4', tcell.ModAlt): action.BgLock,

		RuneEvent('[', tcell.ModAlt): action.Resize,
	}
}

func (m *Editor) HandleEvent(event tcell.Event) {
	if EVENT_MONITORING {
		if ev, ok := event.(*tcell.EventKey); ok {
			m.notification.PushNotification("", fmt.Sprintf("%v", ev.Name()), NotificationNormal)
		}
		if ev, ok := event.(*tcell.EventMouse); ok {
			m.notification.PushNotification("", fmt.Sprintf("%v", ev), NotificationNormal)
		}
	}

	// Events are handled in the following order:
	// - If a modal tool is active, it grabs all non-critical events.
	// - Console resize events will automatically resize the canvas and scale the offset
	// accordingly.
	// - Any kind of panning event (either starting or stopping panning)
	// - Any color picking event (clicking alt and dragging)
	// - If the color selector is active, then any color selection command
	// - Any alt+_ shortcut
	// - If a prompt tool is active, it grabs all incoming events
	// - If the current character pressed is a space or a printable character,
	// then the brush character is set.
	// - Finally it falls through to the current tool if all handlers fail.
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
	case *tcell.EventKey:
		if ev.Key() == tcell.KeyEscape {
			if m.hasModalTool {
				m.ClearModalTool()
				return
			}
			m.ClearTool()
			return
		}
	}

	m.notification.HandleEvent(event)

	if m.hasModalTool {
		m.currentModalTool.HandleEvent(m, event)
		return
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

	if handled := m.HandlePaste(event); handled {
		return
	}
}

func (m *Editor) HandlePaste(event tcell.Event) bool {
	switch ev := event.(type) {
	case *tcell.EventPaste:
		if ev.Start() {
			m.isPasting = true
			m.pendingPasteData = []byte{}
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
			m.pendingPasteData = append(m.pendingPasteData, byte(r))
		default:
			m.pendingPasteData = append(m.pendingPasteData, '\n')
		}
		return true
	}

	return false
}

func (m *Editor) HandlePan(event tcell.Event) bool {
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

func (m *Editor) HandleColorPick(event tcell.Event) bool {
	switch ev := event.(type) {
	case *tcell.EventMouse:
		if ev.Modifiers()&tcell.ModAlt != 0 {
			if m.colorPickState != ColorPickDrag {
				m.colorPickState = ColorPickHover
				canvasX, canvasY := m.cursorX-m.offsetX, m.cursorY-m.offsetY

				curCanvas := m.CurrentCanvas()
				if curCanvas.Data.InBounds(canvasX, canvasY) {
					cell := curCanvas.Data.MustGet(canvasX, canvasY)
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

func (m *Editor) HandleColorSelect(event tcell.Event) bool {
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

func (m *Editor) HandleShortcuts(event tcell.Event) bool {
	switch ev := event.(type) {
	case *tcell.EventKey:
		if act, ok := m.keymap[ParseEvent(ev)]; ok {
			switch act {
			case action.CenterCanvas:
				if m.isPan {
					m.offsetX += m.cursorX - m.panOriginX
					m.offsetY += m.cursorY - m.panOriginY
					m.panOriginX = m.cursorX
					m.panOriginY = m.cursorY
				}
				m.CenterCanvas()

			case action.ForceQuit:
				m.app.WillQuit = true

			case action.Quit:
				if m.HasUnsavedChanges() {
					m.SetModalTool(&YesNoPromptTool{
						prompt:    "File has unsaved changes, quit?",
						yesString: "Save changes and quit",
						yesAction: func() {
							m.SetModalTool(MakePromptTool(
								func(s string) {
									if _, err := m.Save(s); err == nil {
										m.app.WillQuit = true
									}
								},
								"Save to ascii-draw file and quit",
								"save path...",
								m.savedFile,
							))
						},
						noString: "Quit without saving",
						noAction: func() {
							m.app.WillQuit = true
						},
					})
				} else {
					m.app.WillQuit = true
				}

			case action.Help:
				m.SetModalTool(&HelpTool{})

			case action.Save:
				m.SetModalTool(MakePromptTool(
					func(s string) {
						_, _ = m.Save(s)
					},
					"Save to ascii-draw file",
					"save path...",
					m.savedFile,
				))

			case action.Load:
				if m.HasUnsavedChanges() {
					m.SetModalTool(&YesNoPromptTool{
						prompt:    "File has unsaved changes, load file?",
						yesString: "Save changes and load file",
						yesAction: func() {
							m.SetModalTool(MakePromptTool(
								func(s string) {
									if _, err := m.Save(s); err == nil {
										m.SetModalTool(MakePromptTool(
											m.Load,
											"Load binary file",
											"load path...",
											"",
										))
									}
								},
								"Save to ascii-draw file",
								"save path...",
								m.savedFile,
							))
						},
						noString: "Load file without saving current file",
						noAction: func() {
							m.SetModalTool(MakePromptTool(
								m.Load,
								"Load ascii-draw file",
								"load path...",
								"",
							))
						},
					})
				} else {
					m.SetModalTool(MakePromptTool(
						m.Load,
						"Load ascii-draw file",
						"load path...",
						"",
					))
				}

			case action.Export:
				m.SetModalTool(MakePromptTool(
					m.Export,
					"Export to plaintext",
					"export path...",
					"",
				))

			case action.Import:
				if m.HasUnsavedChanges() {
					m.SetModalTool(&YesNoPromptTool{
						prompt:    "File has unsaved changes, import file?",
						yesString: "Save changes and import file",
						yesAction: func() {
							m.SetModalTool(MakePromptTool(
								func(s string) {
									if _, err := m.Save(s); err == nil {
										m.SetModalTool(MakePromptTool(
											m.Import,
											"Import plaintext",
											"import path...",
											"",
										))
									}
								},
								"Save to ascii-draw file",
								"save path...",
								m.savedFile,
							))
						},
						noString: "Import file without saving current file",
						noAction: func() {
							m.SetModalTool(MakePromptTool(
								m.Import,
								"Import plaintext",
								"import path...",
								"",
							))
						},
					})
				} else {
					m.SetModalTool(MakePromptTool(
						m.Import,
						"Import plaintext",
						"import path...",
						"",
					))
				}

			case action.FgColorSelector:
				if m.colorSelectState != ColorSelectFg {
					m.colorSelectState = ColorSelectFg
				} else {
					m.colorSelectState = ColorSelectNone
				}

			case action.BgColorSelector:
				if m.colorSelectState != ColorSelectBg {
					m.colorSelectState = ColorSelectBg
				} else {
					m.colorSelectState = ColorSelectNone
				}

			case action.NewCanvas:
				if m.HasUnsavedChanges() {
					m.SetModalTool(&YesNoPromptTool{
						prompt:    "File has unsaved changes, create new file?",
						yesString: "Save changes and create new file",
						yesAction: func() {
							m.SetModalTool(MakePromptTool(
								func(s string) {
									if _, err := m.Save(s); err == nil {
										m.canvas = MakeBuffer(INIT_WIDTH, INIT_HEIGHT)
										m.savedFile = ""

										m.Reset()
										m.ClearHistory()

										m.cursorX, m.cursorY = m.sw/2, m.sh/2
									}
								},
								"Save to ascii-draw file",
								"save path...",
								m.savedFile,
							))
						},
						noString: "Create new file without saving",
						noAction: func() {
							m.canvas = MakeBuffer(INIT_WIDTH, INIT_HEIGHT)
							m.savedFile = ""

							m.Reset()
							m.ClearHistory()

							m.cursorX, m.cursorY = m.sw/2, m.sh/2
						},
					})
				} else {
					m.canvas = MakeBuffer(INIT_WIDTH, INIT_HEIGHT)
					m.savedFile = ""

					m.Reset()
					m.ClearHistory()

					m.cursorX, m.cursorY = m.sw/2, m.sh/2
				}

			case action.Lasso:
				m.SetTool(&LassoTool{})

			case action.Translate:
				m.SetTool(&TranslateTool{})

			case action.Line:
				m.SetTool(&LineTool{})

			case action.Deselect:
				m.Stage()
				m.stagingCanvas.Deselect()
				m.Commit()

			case action.Copy:
				m.SetClipboard()

			case action.Cut:
				m.SetClipboard()
				m.Stage()
				m.stagingCanvas.ClearSelection()
				m.Commit()

			case action.Paste:
				m.SetTool(&StampTool{})

			case action.Undo:
				m.undoHistoryPos = max(0, m.undoHistoryPos-1)

			case action.Redo:
				m.undoHistoryPos = min(len(m.undoHistory), m.undoHistoryPos+1)

			case action.IncreaseBrushRadius:
				m.brushRadius = min(MAX_BRUSH_RADIUS, m.brushRadius+1)

				// Decrease brush radius
			case action.DecreaseBrushRadius:
				m.brushRadius = max(1, m.brushRadius-1)

				// Resize
			case action.Resize:
				t := &ResizeTool{}
				t.SetDimsFromSelection(m.CurrentCanvas())
				t.InitHandles()
				m.SetTool(t)

				// Toggle alpha lock
			case action.AlphaLock:
				m.lockMask ^= LockMaskAlpha

				// Toggle character lock
			case action.CharLock:
				m.lockMask ^= LockMaskChar

				// Toggle foreground color lock
			case action.FgLock:
				m.lockMask ^= LockMaskFg

				// Toggle background color lock
			case action.BgLock:
				m.lockMask ^= LockMaskBg

				// Clear selection
			case action.ClearSelection:
				m.Stage()
				m.stagingCanvas.ClearSelection()
				m.Commit()

				// Fill selection with current brush
			case action.FillSelection:
				m.Stage()
				c := Cell{
					Value: m.brushCharacter,
					Style: tcell.StyleDefault.Foreground(m.fgColor).Background(m.bgColor),
				}
				m.stagingCanvas.FillSelection(c, m.lockMask)
				m.Commit()
			}
			return true
		}
	}
	return false
}

func (m *Editor) Update() {
	m.notification.Update()
}

func (m *Editor) Draw(p Painter, x, y, w, h int, lag float64) {
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
		t := time.Now().Sub(m.appStartTime).Seconds()
		cx, cy := m.cursorX+m.sx, m.cursorY+m.sy
		cc := m.CurrentCanvas()

		// Draw crosshairs for alignment
		if cy >= m.sy+canvasOffY && cy < m.sy+canvasOffY+cc.Data.Height {
			for x := range cc.Data.Width {
				dt := math.Cos(t*10 - math.Abs(float64(x-(cx-m.sx-canvasOffX)))*1)
				if dt > 0 {
					p.SetByte(
						m.sx+m.offsetX+x,
						cy,
						'.',
						tcell.StyleDefault.Foreground(tcell.ColorWhite),
					)
				}
			}
		}

		if cx >= m.sx+canvasOffX && cx < m.sx+canvasOffX+cc.Data.Width {
			for y := range cc.Data.Height {
				dt := math.Cos(t*6 - math.Abs(float64(y-(cy-m.sy-canvasOffY)))*2)
				if dt > 0 {
					p.SetByte(
						cx,
						m.sy+m.offsetY+y,
						'.',
						tcell.StyleDefault.Foreground(tcell.ColorWhite),
					)
				}
			}
		}

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

	// undo history
	undoHistoryLine := "Already at newest change"
	if m.isStaging {
		undoHistoryLine = "Modification in progress..."
	} else if len(m.undoHistory) > 0 && m.undoHistoryPos == 0 {
		undoHistoryLine = "Already at oldest change"
	} else if m.undoHistoryPos < len(m.undoHistory) {
		undoHistoryLine = fmt.Sprintf("Undo: %d/%d", m.undoHistoryPos+1, len(m.undoHistory)+1)
	}

	SetString(p, x+1, y+m.sh+m.sy, undoHistoryLine, tcell.StyleDefault)

	// current filename
	currentFile := m.savedFile
	unsavedIndicator := ""
	if currentFile == "" {
		currentFile = "New File"
	}
	if m.HasUnsavedChanges() {
		unsavedIndicator = "* "
	}

	fileString := fmt.Sprintf("%s%s", unsavedIndicator, currentFile)

	SetString(
		p,
		x+m.sw-Condition.StringWidth(fileString)+m.sx,
		y+m.sh+m.sy,
		fileString,
		tcell.StyleDefault,
	)

	// modal tool
	if m.hasModalTool {
		m.currentModalTool.Draw(m, p, x, y, w, h, lag)
	}

	m.notification.Draw(p, m.sx, m.sy, m.sw, m.sh, lag)
}

func (m *Editor) DrawStatusBar(p Painter, x, y, w, h int) {
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

func (m *Editor) DrawCanvas(p Painter, offX, offY int) {
	// Rendering the canvas
	curCanvas := m.CurrentCanvas()
	curCanvas.RenderWith(p, offX, offY, true)

	BorderBox(p, Area{
		X:      offX - 1,
		Y:      offY - 1,
		Width:  curCanvas.Data.Width + 2,
		Height: curCanvas.Data.Height + 2,
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

func (m *Editor) ScreenResize(sw, sh int) {
	m.sw, m.sh = sw-2, sh-2
	m.sx, m.sy = 1, 1
}

func (m *Editor) ScaleOffset(oldsw, oldsh, newsw, newsh int) {
	ocx, ocy := m.offsetX-oldsw/2, m.offsetY-oldsh/2
	sfx, sfy := float64(newsw)/float64(oldsw), float64(newsh)/float64(oldsh)
	m.offsetX = int(float64(ocx)*sfx) + newsw/2
	m.offsetY = int(float64(ocy)*sfy) + newsh/2
}

func (m *Editor) ResizeCanvas(newRect Area) {
	curCanvas := m.CurrentCanvas()
	m.Stage()
	m.stagingCanvas = MakeBuffer(newRect.Width, newRect.Height)
	for y := range curCanvas.Data.Height {
		for x := range curCanvas.Data.Width {
			nx, ny := x-newRect.X, y-newRect.Y
			m.stagingCanvas.Data.Set(nx, ny, curCanvas.Data.MustGet(x, y))
			m.stagingCanvas.SelectionMask.Set(nx, ny, curCanvas.SelectionMask.MustGet(x, y))
		}
	}
	m.Commit()
	m.offsetX += newRect.X
	m.offsetY += newRect.Y
	m.ClearTool()
	m.ClearModalTool()
}

func (m *Editor) CenterCanvas() {
	curCanvas := m.CurrentCanvas()
	cw, ch := curCanvas.Data.Width, curCanvas.Data.Height
	m.offsetX = (m.sw - cw) / 2
	m.offsetY = (m.sh - ch) / 2
}

func (m *Editor) SetTool(tool Tool) {
	m.Rollback()
	m.hasTool = true
	m.currentTool = tool
}

func (m *Editor) ClearTool() {
	m.Rollback()
	m.hasTool = true
	m.currentTool = &BrushTool{}
}

func (m *Editor) SetModalTool(tool Tool) {
	m.Rollback()
	m.hasModalTool = true
	m.currentModalTool = tool
}

func (m *Editor) ClearModalTool() {
	m.Rollback()
	m.hasModalTool = false
	m.currentModalTool = nil
}

func (m *Editor) SetFgColor(fg int) {
	if fg == 16 {
		m.fgColor = tcell.ColorDefault
	} else {
		m.fgColor = tcell.ColorValid + tcell.Color(fg)
	}
}

func (m *Editor) SetBgColor(bg int) {
	if bg == 16 {
		m.bgColor = tcell.ColorDefault
	} else {
		m.bgColor = tcell.ColorValid + tcell.Color(bg)
	}
}

func (m *Editor) Export(s string) {
	var msg string
	var err error
	defer func() {
		m.ClearTool()
		m.ClearModalTool()
		if err != nil {
			m.notification.PushNotification("Error", err.Error(), NotificationCritical)
		} else {
			m.notification.PushNotification("", msg, NotificationNormal)
		}
	}()

	if err1 := m.CurrentCanvas().ExportToFile(s); err1 != nil {
		err = err1
		return
	}

	msg = fmt.Sprintf("Successfully exported to plaintext file %s", s)
	m.app.Logger.Printf("Successfully exported to plaintext file %s", s)
}

func (m *Editor) Import(s string) {
	var msg string
	var err error
	defer func() {
		m.ClearTool()
		m.ClearModalTool()
		if err != nil {
			m.notification.PushNotification("Error", err.Error(), NotificationCritical)
		} else {
			m.notification.PushNotification("", msg, NotificationNormal)
		}
	}()

	newCanvas := &Buffer{}

	if err1 := newCanvas.ImportFromFile(s); err1 != nil {
		err = err1
		return
	}

	m.canvas = newCanvas

	m.Reset()
	m.ClearHistory()
	m.savedFile = ""
	m.savedUndoIndex = m.undoHistoryPos
	m.historyChanged = false

	msg = fmt.Sprintf("Successfully imported plaintext file %s", s)
	m.app.Logger.Printf("Successfully imported plaintext file %s", s)
}

// FIXME: this is a massive code smell. maybe make these IO functions error out and have the
// caller deal with them somehow? i guess the rub is that these processses are inherently
// asynchronous, so i might need a special way to take care of that
func (m *Editor) Save(s string) (msg string, err error) {
	defer func() {
		m.ClearTool()
		m.ClearModalTool()
		if err != nil {
			m.notification.PushNotification("Error", err.Error(), NotificationCritical)
		} else {
			m.notification.PushNotification("", msg, NotificationNormal)
		}
	}()

	if m.undoHistoryPos < len(m.undoHistory) && m.CurrentCanvas() == m.canvas {
		panic(1)
	}

	if err1 := m.CurrentCanvas().SaveToFile(s); err1 != nil {
		err = err1
		return
	}
	m.savedFile = s
	m.savedUndoIndex = m.undoHistoryPos
	m.historyChanged = false

	msg = fmt.Sprintf("Successfully saved %s", s)
	m.app.Logger.Printf("Successfully saved binary file %s", s)

	return msg, err
}

func (m *Editor) Load(s string) {
	var msg string
	var err error
	defer func() {
		m.ClearTool()
		m.ClearModalTool()
		if err != nil {
			m.notification.PushNotification("Error", err.Error(), NotificationCritical)
		} else {
			m.notification.PushNotification("", msg, NotificationNormal)
		}
	}()

	newCanvas := &Buffer{}

	if err1 := newCanvas.LoadFromFile(s); err1 != nil {
		err = err1
		return
	}

	m.canvas = newCanvas

	m.Reset()
	m.ClearHistory()
	m.savedFile = s
	m.savedUndoIndex = m.undoHistoryPos
	m.historyChanged = false

	msg = fmt.Sprintf("Successfully loaded %s", s)
	m.app.Logger.Printf("Successfully loaded binary file %s", s)
}

func (m *Editor) SetClipboard() {
	m.clipboard = m.CurrentCanvas().CopySelection()
}

func (m *Editor) SetClipboardFromPasteData() {
	var width, height int
	var w = 0
	for _, c := range m.pendingPasteData {
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
	for _, c := range m.pendingPasteData {
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

func (m *Editor) Stage() {
	if m.isStaging {
		return
	}

	curCanvas := m.CurrentCanvas()
	m.isStaging = true
	m.stagingCanvas = curCanvas.Clone()
}

func (m *Editor) Commit() {
	if !m.isStaging {
		return
	}

	curCanvas := m.canvas
	if m.undoHistoryPos < len(m.undoHistory) {
		curCanvas = m.undoHistory[m.undoHistoryPos]
	}

	if curCanvas.Equal(m.stagingCanvas) {
		m.isStaging = false
		m.stagingCanvas = nil
		return
	}

	if m.undoHistoryPos < len(m.undoHistory) {
		m.undoHistory = m.undoHistory[:m.undoHistoryPos+1]
	} else {
		m.undoHistory = append(m.undoHistory, m.canvas)
	}

	if m.savedUndoIndex > m.undoHistoryPos {
		m.historyChanged = true
	}
	m.undoHistoryPos++
	m.isStaging = false
	m.canvas = m.stagingCanvas
	m.stagingCanvas = nil

	m.app.Logger.Println("Committed new action to canvas")
}

func (m *Editor) Rollback() {
	if !m.isStaging {
		return
	}
	m.isStaging = false
	m.stagingCanvas = nil
}

func (m *Editor) CurrentCanvas() *Buffer {
	if m.isStaging {
		return m.stagingCanvas
	} else if m.undoHistoryPos < len(m.undoHistory) {
		return m.undoHistory[m.undoHistoryPos]
	} else {
		return m.canvas
	}
}

func (m *Editor) IsPaintTool() bool {
	if _, ok := m.currentTool.(*BrushTool); ok {
		return true
	}
	if _, ok := m.currentTool.(*LineTool); ok {
		return true
	}
	return false
}

func (m *Editor) HasUnsavedChanges() bool {
	return m.historyChanged || m.undoHistoryPos != m.savedUndoIndex
}

func (m *Editor) Reset() {
	m.Rollback()
	m.ClearTool()
	m.ClearModalTool()
	m.CenterCanvas()
	m.isPan = false
	m.colorPickState = ColorPickNone
	m.colorSelectState = ColorSelectNone
	m.historyChanged = false
	m.savedUndoIndex = 0
	m.lockMask = 0
}

func (m *Editor) ClearHistory() {
	m.undoHistoryPos = 0
	m.undoHistory = m.undoHistory[:0]
}
