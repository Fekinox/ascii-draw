package main

import "github.com/gdamore/tcell/v2"

type Tool interface {
	HandleEvent(m *Editor, event tcell.Event)
	Draw(m *Editor, p Painter, x, y, w, h int, lag float64)
}
