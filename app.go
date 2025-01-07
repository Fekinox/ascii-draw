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

	keyActionMap  map[tcell.Key]Action
	runeActionMap map[rune]Action

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
		keyActionMap:  make(map[tcell.Key]Action),
		runeActionMap: make(map[rune]Action),
	}

	app.keyActionMap[tcell.KeyLeft] = MoveLeft
	app.keyActionMap[tcell.KeyRight] = MoveRight
	app.keyActionMap[tcell.KeyUp] = MoveUp
	app.keyActionMap[tcell.KeyDown] = MoveDown
	app.keyActionMap[tcell.KeyEnter] = MenuConfirm

	app.runeActionMap['q'] = Quit
	app.runeActionMap['Q'] = Quit

	// Initialize logger
	app.LogFileHandle, err = os.Create("logfile")
	if err != nil {
		log.Fatalf("%+v", err)
	}

	app.Logger = log.New(app.LogFileHandle, "", log.Flags())

	app.widget = NewMultiWidget(
		Init(),
		&Monitor{},
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
			case *tcell.EventKey:
				if ev.Key() == tcell.KeyEscape || ev.Key() == tcell.KeyCtrlC {
					return
				} else if ev.Key() == tcell.KeyCtrlL {
					Screen.Sync()
				} else {
					var action Action
					var ok bool
					if ev.Key() == tcell.KeyRune {
						action, ok = a.runeActionMap[ev.Rune()]
					} else {
						action, ok = a.keyActionMap[ev.Key()]
					}

					// action, ok
					var _, _ = action, ok
					if ok {
						a.widget.HandleAction(action)
					} else {
						a.widget.HandleEvent(ev)
					}
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
	tw, th := sw-2, sh-2

	if tw < MIN_WIDTH || th < MIN_HEIGHT {
		ShowResizeScreen(tw, th, defStyle)
		Screen.Show()
		return
	}

	rr := Area{
		X:      (sw - tw) / 2,
		Y:      (sh - th) / 2,
		Width:  tw,
		Height: th,
	}

	BorderBox(Area{
		X:      rr.X - 1,
		Y:      rr.Y - 1,
		Width:  rr.Width + 1,
		Height: rr.Height + 1,
	}, defStyle)

	a.widget.Draw(Screen, rr.X, rr.Y, rr.Width, rr.Height, lag)

	Screen.Show()
}
