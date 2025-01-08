package main

import (
	"github.com/gdamore/tcell/v2"
)

type PromptTool struct {
	Text TextWidget
}

func MakePromptTool(onSubmit func(s string)) *PromptTool {
	t := &PromptTool{
		Text: TextWidget{
			Active: true,
			Hint:   "export path...",
		},
	}

	t.Text.OnSubmit = onSubmit
	return t
}

func (e *PromptTool) Name() string {
	return "Export"
}

func (e *PromptTool) HandleAction(m *MainWidget, action Action) {
	e.Text.HandleAction(action)
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
