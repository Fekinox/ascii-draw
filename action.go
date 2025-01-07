package main

type Action int8

const (
	MoveUp Action = iota
	MoveDown
	MoveLeft
	MoveRight
	Quit
	MenuConfirm
)

var ActionNames = []string{
	"MoveUp",
	"MoveDown",
	"MoveLeft",
	"MoveRight",
	"Quit",
	"MenuConfirm",
}

type ReplayAction struct {
	Action Action
	Frame  int64
}

func (a Action) ToString() string {
	return ActionNames[a]
}
