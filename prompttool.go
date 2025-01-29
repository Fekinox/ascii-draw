package main

import (
	"github.com/gdamore/tcell/v2"
)

type PromptTool struct {
	Text   TextWidget
	Prompt string
}

func MakePromptTool(onSubmit func(s string), prompt, hint, initContents string) *PromptTool {
	t := &PromptTool{
		Text: TextWidget{
			Active: true,
			Hint:   hint,
		},
		Prompt: prompt,
	}

	t.Text.OnSubmit = onSubmit
	t.Text.SetContents(initContents)
	return t
}

func (e *PromptTool) Name() string {
	return "Export"
}

func (e *PromptTool) HandleEvent(m *MainWidget, event tcell.Event) {
	e.Text.HandleEvent(event)
}

func (e *PromptTool) Update(m *MainWidget) {
	e.Text.Update()
}

func (e *PromptTool) Draw(
	m *MainWidget,
	p Painter,
	x, y, w, h int,
	lag float64,
) {
	r := Area{
		Width:  50,
		Height: 2,
	}
	r.X = x + (w-r.Width)/2
	r.Y = y + (h-r.Height)/2
	crop := &CropPainter{
		p:    p,
		area: r,
	}
	bb := Area{
		X:      r.X - 1,
		Y:      r.Y - 1,
		Width:  r.Width + 2,
		Height: r.Height + 2,
	}
	BorderBox(p, bb, tcell.StyleDefault)
	FillRegion(p, r.X, r.Y, r.Width, r.Height, ' ', tcell.StyleDefault)
	SetString(p, r.X, r.Y, e.Prompt, tcell.StyleDefault)
	e.Text.Draw(crop, r.X, r.Y+1, r.Width, r.Height, lag)
}
