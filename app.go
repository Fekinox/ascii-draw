package main

import (
	"errors"
	"log"
	"os"
	"time"

	"github.com/gdamore/tcell/v2"
)

const FRAMES_PER_SECOND int64 = 60
const UPDATE_TICK_RATE_MS float64 = 1000.0 / float64(FRAMES_PER_SECOND)
const TIME_SCALE float64 = 1

type App struct {
	widget             Widget
	lastRenderDuration float64
	DefaultStyle       tcell.Style

	WillQuit bool

	LogFileHandle *os.File
	Logger        *log.Logger
}

func NewApp() *App {
	s, err := tcell.NewScreen()
	if err != nil {
		log.Fatalf("%+v", err)
	}
	Screen = s
	if err := Screen.Init(); err != nil {
		log.Fatalf("%+v", err)
	}

	Screen.SetStyle(defStyle)
	Screen.EnableMouse()
	Screen.EnablePaste()
	Screen.Clear()

	if !Screen.HasMouse() {
		log.Fatalf("%+v", errors.New("no mouse"))
	}

	app := &App{
		DefaultStyle: tcell.StyleDefault.Background(tcell.ColorReset).
			Foreground(tcell.ColorReset),
	}

	// Initialize logger
	app.LogFileHandle, err = os.Create("logfile")
	if err != nil {
		log.Fatalf("%+v", err)
	}

	app.Logger = log.New(app.LogFileHandle, "", log.Flags())

	app.widget = NewMultiWidget(
		Init(app, Screen),
		// &Monitor{},
	)

	return app
}

func (a *App) Quit() {
	maybePanic := recover()
	Screen.Fini()
	if maybePanic != nil {
		panic(maybePanic)
	}
}

func (a *App) Loop() {
	lag := 0.0
	prevTime := time.Now()

	for {
		currTime := time.Now()
		elapsed := float64(currTime.Sub(prevTime).Nanoseconds()) / (1000 * 1000)
		lag += elapsed * TIME_SCALE
		prevTime = currTime

		if a.WillQuit {
			return
		}

		// Event handling
		for Screen.HasPendingEvent() {
			ev := Screen.PollEvent()
			switch ev := ev.(type) {
			case *tcell.EventResize:
				Screen.Sync()
				a.widget.HandleEvent(ev)
			case *tcell.EventKey:
				if ev.Key() == tcell.KeyCtrlL {
					Screen.Sync()
				} else {
					a.widget.HandleEvent(ev)
				}
			default:
				a.widget.HandleEvent(ev)
			}
		}

		dirty := false
		for lag >= UPDATE_TICK_RATE_MS {
			dirty = true
			a.widget.Update()
			lag -= UPDATE_TICK_RATE_MS
		}

		if dirty {
			a.Draw(lag)
		}
	}
}

func (a *App) Draw(lag float64) {
	Screen.Clear()

	sw, sh := Screen.Size()

	p := Paint

	if sw < MIN_WIDTH || sh < MIN_HEIGHT {
		ShowResizeScreen(p, sw, sh, defStyle)
		Screen.Show()
		return
	}

	a.widget.Draw(p, 0, 0, sw, sh, lag)

	Screen.Show()
}
