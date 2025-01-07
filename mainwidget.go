package main

import (
	"github.com/gdamore/tcell/v2"
)

type MainWidget struct {
	sw int
	sh int

	cursorX int
	cursorY int

	selectionMask Grid[bool]

	currentState EditorState
}

type LineSegment struct {
	start Position
	end   Position
}

var (
	_ Widget = &MainWidget{}
)

func Init() *MainWidget {
	w := &MainWidget{
		selectionMask: MakeGrid(200, 200, false),
		currentState:  &NormalState{},
	}
	w.currentState.OnEnter(w)
	return w
}

func (m *MainWidget) HandleAction(action Action) {
	switch action {
	case MoveUp:
		m.cursorY--
	case MoveDown:
		m.cursorY++
	case MoveLeft:
		m.cursorX--
	case MoveRight:
		m.cursorX++
	}

	m.currentState.HandleAction(m, action)
}

func (m *MainWidget) HandleEvent(event tcell.Event) {
	switch ev := event.(type) {
	case *tcell.EventResize:
		m.sw, m.sh = ev.Size()
	case *tcell.EventMouse:
		cx, cy := ev.Position()
		m.cursorX, m.cursorY = cx, cy
	}

	m.currentState.HandleEvent(m, event)
}

func (m *MainWidget) Update() {
	m.currentState.Update(m)
}

func (m *MainWidget) Draw(screen tcell.Screen, x, y, w, h int, lag float64) {
	// screen.SetContent(m.cursorX, m.cursorY, '#', nil, tcell.StyleDefault)
	for yy := range m.selectionMask.Height {
		for xx := range m.selectionMask.Width {
			if v, ok := m.selectionMask.Get(x+xx, y+yy); ok && v {
				screen.SetContent(x+xx, x+yy, '.', nil, tcell.StyleDefault)
			}
		}
	}
	m.currentState.Draw(m, screen, x, y, w, h, lag)
}

func (m *MainWidget) SetState(state EditorState) {
	m.currentState = state
	m.currentState.OnEnter(m)
}
