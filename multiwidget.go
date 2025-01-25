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

func (m *MultiWidget) Draw(p Painter, x, y, w, h int, lag float64) {
	for _, wd := range m.widgets {
		wd.Draw(p, x, y, w, h, lag)
	}
}
