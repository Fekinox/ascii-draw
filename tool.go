package main

import "github.com/gdamore/tcell/v2"

type Tool interface {
	HandleEvent(m *MainWidget, event tcell.Event)
	Draw(m *MainWidget, p Painter, x, y, w, h int, lag float64)
}
