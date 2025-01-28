package main

import (
	"strings"
	"unicode"

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
			var i int
			for _, r := range t.Contents {
				if i == t.Cursor {
					i++
					continue
				}
				sb.WriteRune(r)
				i++
			}
			t.Contents = sb.String()
		case tcell.KeyLeft:
			t.Cursor = max(0, t.Cursor-1)
		case tcell.KeyRight:
			t.Cursor = min(len(t.Contents), t.Cursor+1)
		case tcell.KeyEnter:
			t.OnSubmit(t.Contents)
		case tcell.KeyRune:
			r := ev.Rune()
			if r != ' ' && !unicode.IsGraphic(r) {
				return
			}
			var sb strings.Builder
			var done bool
			var i int
			for _, r := range t.Contents {
				if i == t.Cursor {
					done = true
					sb.WriteRune(r)
				}
				sb.WriteRune(r)
				i++
			}
			if !done {
				sb.WriteRune(r)
			}
			t.Contents = sb.String()
			t.Cursor += 1
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
	var i, col int
	for _, r := range t.Contents {
		if i == t.Cursor {
			break
		}
		col += Condition.RuneWidth(r)
		i++
	}
	p.SetStyle(x+1+col, y, tcell.StyleDefault.Reverse(true))
}
