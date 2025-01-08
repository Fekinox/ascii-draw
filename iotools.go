package main

import (
	"os"

	"github.com/gdamore/tcell/v2"
)

type SaveTool struct {
}

type LoadTool struct {
}

type ExportTool struct {
	m    *MainWidget
	Text TextWidget
}

type ImportTool struct {
	Text TextWidget
}

func MakeExportTool(m *MainWidget) *ExportTool {
	t := &ExportTool{
		Text: TextWidget{
			Active: true,
			Hint:   "export path...",
		},
		m: m,
	}

	t.Text.OnSubmit = t.Export

	return t
}

func (e *ExportTool) Name() string {
	return "Export"
}

func (e *ExportTool) HandleAction(m *MainWidget, action Action) {
	e.Text.HandleAction(action)
}

func (e *ExportTool) HandleEvent(m *MainWidget, event tcell.Event) {
	e.Text.HandleEvent(event)
}

func (e *ExportTool) Update(m *MainWidget) {
	e.Text.Update()
}

func (e *ExportTool) Draw(
	m *MainWidget,
	p Painter,
	x, y, w, h int,
	lag float64,
) {
	e.Text.Draw(p, x, y, w, h, lag)
}

func (e *ExportTool) Export(s string) {
	f, err := os.Create(s)
	if err != nil {
		panic(err)
	}
	e.m.canvas.Export(f)
}
