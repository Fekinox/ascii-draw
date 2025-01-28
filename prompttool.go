package main

import (
	"github.com/gdamore/tcell/v2"
)

type PromptTool struct {
	Text TextWidget
}

func MakePromptTool(onSubmit func(s string), hint, initContents string) *PromptTool {
	t := &PromptTool{
		Text: TextWidget{
			Active: true,
			Hint:   hint,
		},
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
	e.Text.Draw(p, x, y, w, h, lag)
}
