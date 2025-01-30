package main

import (
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
)

const NOTIFICATION_WIDGET_MAX_WIDTH int = 40

type NotificationWidget struct {
	active    bool
	priority  int
	startTime time.Time
	header    []string
	body      []string
	sx        int
	sy        int
	sw        int
	sh        int
}

// Greedy word wrapping algorithm
func wrap(s string, width int) []string {
	var res []string
	var builder strings.Builder
	var currentWidth int
	for _, c := range s {
		w := Condition.RuneWidth(c)
		if currentWidth+w > width {
			currentWidth = w
			res = append(res, builder.String())
			builder.Reset()
			builder.WriteRune(c)
		} else {
			builder.WriteRune(c)
			currentWidth += w
		}
	}
	if builder.Len() > 0 {
		res = append(res, builder.String())
	}
	return res
}

var (
	_ Widget = &NotificationWidget{}
)

func (n *NotificationWidget) HandleEvent(event tcell.Event) {
	switch ev := event.(type) {
	case *tcell.EventMouse:
		cx, cy := ev.Position()
		cx, cy = cx-n.sx, cy-n.sy
		if ev.Buttons()&tcell.Button1 != 0 {
			w, h := NOTIFICATION_WIDGET_MAX_WIDTH, len(n.header)+len(n.body)
			r := Area{
				X:      n.sx + (n.sw-w)/2 - 1,
				Y:      n.sy,
				Width:  w + 2,
				Height: h + 2,
			}
			if r.Contains(cx, cy) {
				n.active = false
			}
		}
	}
}

func (n *NotificationWidget) Update() {
	if n.active && time.Now().Sub(n.startTime).Seconds() > 10 {
		n.active = false
	}
}

func (n *NotificationWidget) Draw(p Painter, x, y, w, h int, lag float64) {
	n.sx, n.sy, n.sw, n.sh = x, y, w, h
	if !n.active {
		return
	}

	r := Area{
		X:      x,
		Y:      y + 1,
		Width:  NOTIFICATION_WIDGET_MAX_WIDTH,
		Height: len(n.header) + len(n.body),
	}

	r.X += (w - r.Width) / 2

	bb := Area{
		X:      r.X - 1,
		Y:      r.Y - 1,
		Width:  r.Width + 2,
		Height: r.Height + 2,
	}

	BorderBox(p, bb, tcell.StyleDefault)
	for i, l := range n.header {
		SetString(p, r.X, r.Y+i, l, tcell.StyleDefault.Foreground(tcell.ColorGreen))
	}
	for i, l := range n.body {
		SetString(p, r.X, r.Y+i+len(n.header), l, tcell.StyleDefault)
	}
}

func (n *NotificationWidget) PushNotification(header string, body string) {
	n.header = wrap(header, NOTIFICATION_WIDGET_MAX_WIDTH)
	n.body = wrap(body, NOTIFICATION_WIDGET_MAX_WIDTH)
	n.active = true
	n.startTime = time.Now()
}
