package main

import (
	"strings"

	"github.com/gdamore/tcell/v2"
)

type TextWidget struct {
	Contents string
	Cursor   int
	Hint     string
	Active   bool

	OnSubmit func(s string)
}

var (
	_ Widget = &TextWidget{}
)

func (t *TextWidget) HandleAction(action Action) {
	if !t.Active {
		return
	}
}

func (t *TextWidget) HandleEvent(event tcell.Event) {
	if !t.Active {
		return
	}

	switch ev := event.(type) {
	case *tcell.EventKey:
		switch ev.Key() {
		case tcell.KeyBackspace2:
			if t.Cursor == 0 {
				return
			}
			t.Cursor--
			var sb strings.Builder
			sb.WriteString(t.Contents[:t.Cursor])
			sb.WriteString(t.Contents[t.Cursor+1:])
			s := sb.String()
			t.Contents = s
		case tcell.KeyLeft:
			t.Cursor = max(0, t.Cursor-1)
		case tcell.KeyRight:
			t.Cursor = min(len(t.Contents), t.Cursor+1)
		case tcell.KeyEnter:
			t.OnSubmit(t.Contents)
		case tcell.KeyRune:
			var sb strings.Builder
			sb.WriteString(t.Contents[:t.Cursor])
			sb.WriteRune(ev.Rune())
			sb.WriteString(t.Contents[t.Cursor:])
			t.Contents = sb.String()
			t.Cursor++
		}
	}
}

func (t *TextWidget) Update() {
	if !t.Active {
		return
	}
}

func (t *TextWidget) Draw(p Painter, x, y, w, h int, lag float64) {
	if t.Contents == "" {
		st := tcell.StyleDefault.Foreground(tcell.ColorGray)
		SetString(p, x+1, y, t.Hint, st)
	} else {
		SetString(p, x+1, y, t.Contents, tcell.StyleDefault)
	}
	p.SetStyle(x+1+t.Cursor, y, tcell.StyleDefault.Reverse(true))
}
