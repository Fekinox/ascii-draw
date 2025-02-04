package action

type Action int32

const (
	Save Action = iota
	Load
	Import
	Export
	Reset
	CenterCanvas
	Quit
	ForceQuit
	Help
	NewCanvas
	FgColorSelector
	BgColorSelector
	Lasso
	Translate
	Deselect
	Copy
	Cut
	Paste
	Undo
	Redo
	IncreaseBrushRadius
	DecreaseBrushRadius
	Resize
	AlphaLock
	CharLock
	FgLock
	BgLock
	ClearSelection
	FillSelection
)
