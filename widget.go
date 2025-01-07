package main

import "github.com/gdamore/tcell/v2"

type Widget interface {
	HandleAction(action Action)
	HandleEvent(event tcell.Event)
	Update()
	Draw(p Painter, x, y, w, h int, lag float64)
}

type ToggleableWidget struct {
	disabled bool
	Widget
}

func (t *ToggleableWidget) HandleAction(action Action) {
	t.Widget.HandleAction(action)
}

func (t *ToggleableWidget) HandleEvent(event tcell.Event) {
	t.Widget.HandleEvent(event)
}

func (t *ToggleableWidget) Update() {
	if !t.disabled {
		t.Widget.Update()
	}
}

func (t *ToggleableWidget) Draw(p Painter, x, y, w, h int, lag float64) {
	if !t.disabled {
		t.Widget.Draw(p, x, y, w, h, lag)
	}
}
