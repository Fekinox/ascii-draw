package main

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
)

type YesNoPromptTool struct {
	prompt    string
	yesString string
	yesAction func()
	noString  string
	noAction  func()
}

func (e *YesNoPromptTool) Name() string {
	return "Quit"
}

func (e *YesNoPromptTool) HandleEvent(m *Editor, event tcell.Event) {
	switch ev := event.(type) {
	case *tcell.EventKey:
		if ev.Key() == tcell.KeyRune {
			switch ev.Rune() {
			case 'y':
				m.ClearModalTool()
				e.yesAction()
			case 'n':
				m.ClearModalTool()
				e.noAction()
			}
		}
	}
}

func (e *YesNoPromptTool) Update(m *Editor) {
}

func (e *YesNoPromptTool) Draw(
	m *Editor,
	p Painter,
	x, y, w, h int,
	lag float64,
) {
	r := Area{
		Width:  50,
		Height: 4,
	}
	r.X = x + (w-r.Width)/2
	r.Y = y + (h-r.Height)/2
	bb := Area{
		X:      r.X - 1,
		Y:      r.Y - 1,
		Width:  r.Width + 2,
		Height: r.Height + 2,
	}
	BorderBox(p, bb, tcell.StyleDefault)
	FillRegion(p, r.X, r.Y, r.Width, r.Height, ' ', tcell.StyleDefault)
	SetString(p, r.X, r.Y, e.prompt, tcell.StyleDefault)
	SetString(p, r.X, r.Y+1, fmt.Sprintf("(y) %s", e.yesString), tcell.StyleDefault)
	SetString(p, r.X, r.Y+2, fmt.Sprintf("(n) %s", e.noString), tcell.StyleDefault)
	SetString(p, r.X, r.Y+3, "(esc) Cancel", tcell.StyleDefault)
}
