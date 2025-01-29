package main

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"
)

type TextWidget struct {
	Contents   string
	Cursor     int
	Hint       string
	Active     bool
	StartIndex int

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
	t.EnsureCursorVisible(w)
	if t.Contents == "" {
		st := tcell.StyleDefault.Foreground(tcell.ColorGray)
		SetString(p, x, y, t.Hint, st)
	} else {
		s := []rune(t.Contents)[t.StartIndex:]
		SetString(p, x, y, string(s), tcell.StyleDefault)
	}
	var col int
	for i, r := range []rune(t.Contents)[t.StartIndex:] {
		if i == t.Cursor-t.StartIndex {
			break
		}
		col += Condition.RuneWidth(r)
		i++
	}
	p.SetStyle(x+col, y, tcell.StyleDefault.Reverse(true))
}

func (t *TextWidget) SetContents(s string) {
	t.Contents = s
	t.Cursor = utf8.RuneCountInString(s)
}

func (t *TextWidget) EnsureCursorVisible(w int) {
	// If cursor is less than the start index, then set the start index to the cursor
	if t.Cursor < t.StartIndex {
		t.StartIndex = t.Cursor
		return
	}
	s := []rune(t.Contents)
	currentWidth := Condition.StringWidth(string(s[t.StartIndex:t.Cursor])) + 1
	for currentWidth > w {
		currentWidth -= Condition.RuneWidth(s[t.StartIndex])
		t.StartIndex++
	}
}
