package main

import "github.com/gdamore/tcell/v2"

type MultiWidget struct {
	widgets []Widget
}

func NewMultiWidget(widgets ...Widget) *MultiWidget {
	ws := make([]Widget, len(widgets))
	for i, w := range widgets {
		ws[i] = w
	}

	return &MultiWidget{widgets: ws}
}

func (m *MultiWidget) HandleAction(action Action) {
	for _, w := range m.widgets {
		w.HandleAction(action)
	}
}

func (m *MultiWidget) HandleEvent(event tcell.Event) {
	for _, w := range m.widgets {
		w.HandleEvent(event)
	}
}

func (m *MultiWidget) Update() {
	for _, w := range m.widgets {
		w.Update()
	}
}

func (m *MultiWidget) Draw(screen tcell.Screen, x, y, w, h int, lag float64) {
	for _, wd := range m.widgets {
		wd.Draw(screen, x, y, w, h, lag)
	}
}
