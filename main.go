package main

import (
	"fmt"
	"runtime/debug"

	"github.com/gdamore/tcell/v2"
)

const MIN_WIDTH = 80
const MIN_HEIGHT = 24

var (
	defStyle tcell.Style
)

func main() {
	if info, ok := debug.ReadBuildInfo(); ok {
		fmt.Println(info.Settings)
	}
	a := NewApp()
	defer a.Quit()
	a.Loop()
}
