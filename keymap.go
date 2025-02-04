package main

import "github.com/gdamore/tcell/v2"

// Key events formatted in a more consistent, non-redundant matter.
type KeyEvent struct {
	Modifiers tcell.ModMask
	Key       tcell.Key
	Rune      rune
}

func ParseEvent(ev *tcell.EventKey) KeyEvent {
	newMod := ev.Modifiers()
	key := ev.Key()
	rune := ev.Rune()
	if key >= tcell.KeyCtrlSpace && key <= tcell.KeyCtrlUnderscore {
		newMod &^= tcell.ModCtrl
	}

	if key != tcell.KeyRune {
		rune = 0
	}
	return KeyEvent{
		Modifiers: newMod,
		Key:       key,
		Rune:      rune,
	}
}

func RuneEvent(r rune, mod tcell.ModMask) KeyEvent {
	return KeyEvent{
		Modifiers: mod,
		Key:       tcell.KeyRune,
		Rune:      r,
	}
}
