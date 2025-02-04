package main

import "github.com/gdamore/tcell/v2"

type HelpTool struct {
	currentPage int
}

var (
	_ Tool = &HelpTool{}
)

var helpPages [][][]string = [][][]string{
	{
		{
			"press any key to set brush char",
			"esc: return to brush tool",
			"ctrl+h: show this help page",
			"ctrl+q: quit",
			"ctrl+alt+Q: force quit",
			"ctrl+f: select fg color",
			"ctrl+g: select bg color",
			"ctrl+n: clear canvas",
			"ctrl+click: pan",
			"alt+=: increase brush radius",
			"alt+-: decrease brush radius",
			"alt+hover: lookup color on canvas",
			"alt+click: grab character",
			"alt+drag up: grab fg color",
			"alt+drag down: grab bg color",
		},
		{
			"alt+1: toggle alpha lock",
			"alt+2: toggle char lock",
			"alt+3: toggle fg lock",
			"alt+4: toggle bg lock",
			"ctrl+s: save to file",
			"ctrl+l: load to file",
			"ctrl+o: import text",
			"ctrl+p: export text",
			"ctrl+z: undo",
			"ctrl+alt+z: redo",
			"ctrl+c: copy",
			"ctrl+x: cut",
			"ctrl+v: paste",
			"ctrl+a: reset selection",
			"alt+,: clear selection",
			"alt+.: fill selection",
		},
	},
	{
		{
			"brush (default tool)",
			"click and drag",
			"tab to toggle straight line mode",
			"",
			"lasso (ctrl+r)",
			"click and drag to make freeform selection",
			"",
			"translate (ctrl+t)",
			"click and drag to move selected characters",
			"",
			"resize (alt+[)",
			"click and drag to set new canvas dimensions",
			"enter to commit",
			"",
		},
		{},
	},
}

func (e *HelpTool) HandleEvent(m *Editor, event tcell.Event) {
	switch ev := event.(type) {
	case *tcell.EventKey:
		if ev.Key() == tcell.KeyTAB {
			e.currentPage = (e.currentPage + 1) % len(helpPages)
		}
	}
}

func (e *HelpTool) Draw(m *Editor, p Painter, x, y, w, h int, lag float64) {
	r := Area{
		Width:  70,
		Height: 20,
	}
	r.X = x + (w-r.Width)/2
	r.Y = y + (h-r.Height)/2
	bb := Area{
		X:      r.X - 1,
		Y:      r.Y - 1,
		Width:  r.Width + 2,
		Height: r.Height + 2,
	}
	BorderBox(p, bb, tcell.StyleDefault)
	FillRegion(p, r.X, r.Y, r.Width, r.Height, ' ', tcell.StyleDefault)

	SetCenteredString(p, r.X+r.Width/2, r.Y, ProgramName(), tcell.StyleDefault)
	SetCenteredString(p, r.X+r.Width/2, r.Y+r.Height, "press tab for next page", tcell.StyleDefault)

	for y, ln := range helpPages[e.currentPage][0] {
		SetString(p, r.X, r.Y+2+y, ln, tcell.StyleDefault)
	}

	for y, ln := range helpPages[e.currentPage][1] {
		SetString(p, r.X+r.Width/2, r.Y+2+y, ln, tcell.StyleDefault)
	}
}
