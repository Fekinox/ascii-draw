package main

import "github.com/gdamore/tcell/v2"

type KeyEvent struct {
	Modifiers tcell.ModMask
	Key       tcell.Key
	Rune      rune
}
