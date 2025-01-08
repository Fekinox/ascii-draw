package main

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
)

type Monitor struct {
	lastAction    Action
	hasLastAction bool
	lastEvent     tcell.Event
	hasLastEvent  bool
}

func (m *Monitor) HandleAction(action Action) {
	m.lastAction = action
	m.hasLastAction = true
}

func (m *Monitor) HandleEvent(event tcell.Event) {
	m.lastEvent = event
	m.hasLastEvent = true
}

func (m *Monitor) Update() {
}

func (m *Monitor) Draw(p Painter, x, y, w, h int, lag float64) {
	if m.hasLastAction {
		SetString(
			p,
			x,
			y+3,
			fmt.Sprintf("Action: %v", m.lastAction),
			tcell.StyleDefault,
		)
	}
	if m.hasLastEvent {
		SetString(
			p,
			x,
			y+4,
			fmt.Sprintf("Event: %v", m.lastEvent),
			tcell.StyleDefault,
		)
	}
}
