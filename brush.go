package main

import "github.com/gdamore/tcell/v2"

type BrushTool struct {
	currentIcon byte
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

			m.canvas.Set(cx, cy, b.currentIcon, tcell.StyleDefault)
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
	p.SetRune(
		m.cursorX,
		m.cursorY,
		rune(b.currentIcon),
		nil,
		tcell.StyleDefault,
	)
}
