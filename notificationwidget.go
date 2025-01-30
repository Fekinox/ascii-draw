package main

import (
	"strings"
	"time"
	"unicode"

	"github.com/gdamore/tcell/v2"
)

const NOTIFICATION_WIDGET_MAX_WIDTH int = 40

type NotificationPriority int

const (
	NotificationNormal NotificationPriority = iota
	NotificationWarning
	NotificationCritical
)

type NotificationWidget struct {
	active    bool
	priority  NotificationPriority
	startTime time.Time
	header    []string
	body      []string
	sx        int
	sy        int
	sw        int
	sh        int
}

// Greedy word wrapping algorithm
func greedyWordWrap(s string, width int) []string {
	if s == "" {
		return []string{}
	}
	if Condition.StringWidth(s) <= width {
		return []string{s}
	}

	var res []string
	var lineBuilder strings.Builder
	var wordBuilder strings.Builder
	var currentWidth int
	var wordWidth int
	for _, c := range s {
		if !unicode.IsSpace(c) {
			wordWidth += Condition.RuneWidth(c)
			wordBuilder.WriteRune(c)
			continue
		}
		w := wordBuilder.String()
		wordBuilder.Reset()
		// if we have a space, mark end of word
		if currentWidth+wordWidth > width {
			res = append(res, lineBuilder.String())
			lineBuilder.Reset()
			currentWidth = 0
		}
		lineBuilder.WriteString(w)
		currentWidth += wordWidth
		wordWidth = 0

		if currentWidth+1 > width {
			res = append(res, lineBuilder.String())
			lineBuilder.Reset()
			currentWidth = 0
		} else {
			lineBuilder.WriteRune(c)
			currentWidth += Condition.RuneWidth(c)
		}
	}
	if wordBuilder.Len() > 0 {
		w := wordBuilder.String()
		if currentWidth+wordWidth > width {
			res = append(res, lineBuilder.String())
			lineBuilder.Reset()
			currentWidth = 0
		}
		lineBuilder.WriteString(w)
		currentWidth += wordWidth + 1
		wordWidth = 0
	}
	if lineBuilder.Len() > 0 {
		res = append(res, lineBuilder.String())
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
		st := tcell.StyleDefault
		switch n.priority {
		case NotificationWarning:
			st = st.Foreground(tcell.ColorYellow)
		case NotificationCritical:
			st = st.Foreground(tcell.ColorRed)
		}
		SetString(p, r.X, r.Y+i, l, st)
	}
	for i, l := range n.body {
		SetString(p, r.X, r.Y+i+len(n.header), l, tcell.StyleDefault)
	}
}

func (n *NotificationWidget) PushNotification(header string, body string, p NotificationPriority) {
	n.header = greedyWordWrap(header, NOTIFICATION_WIDGET_MAX_WIDTH)
	n.body = greedyWordWrap(body, NOTIFICATION_WIDGET_MAX_WIDTH)
	n.active = true
	n.priority = p
	n.startTime = time.Now()
}
