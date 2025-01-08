package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"github.com/gdamore/tcell/v2"
)

type CellColor byte

const magicNumber int64 = 0xdeadbeef

const (
	DefaultColor = 0
	ValidColor   = 1 << 4
)

const (
	ColorBlack = ValidColor + iota
	ColorMaroon
	ColorGreen
	ColorOlive
	ColorNavy
	ColorPurple
	ColorTeal
	ColorSilver
	ColorGray
	ColorRed
	ColorLime
	ColorYellow
	ColorBlue
	ColorFuchsia
	ColorAqua
	ColorWhite
)

type Cell struct {
	Value byte
	Style tcell.Style
}

type Buffer struct {
	Data Grid[Cell]
}

func MakeBuffer(width, height int) *Buffer {
	b := &Buffer{
		Data: MakeGrid(width, height, Cell{Value: ' '}),
	}
	return b
}

func (b *Buffer) Render(screen tcell.Screen, x, y int, overwrite bool) {
	for dy := range b.Data.Height {
		for dx := range b.Data.Width {
			xx, yy := x+dx, y+dy
			c := b.Data.MustGet(dx, dy)
			if overwrite && c.Value == 0 {
				screen.SetContent(xx, yy, ' ', nil, c.Style)
			} else if c.Value != 0 {
				screen.SetContent(xx, yy, rune(c.Value), nil, c.Style)
			}
		}
	}
}

func (b *Buffer) RenderWith(p Painter, x, y int, overwrite bool) {
	for dy := range b.Data.Height {
		for dx := range b.Data.Width {
			xx, yy := x+dx, y+dy
			c := b.Data.MustGet(dx, dy)
			if overwrite && c.Value == 0 {
				p.SetByte(xx, yy, ' ', c.Style)
			} else if c.Value != 0 {
				p.SetByte(xx, yy, c.Value, c.Style)
			}
		}
	}
}

func (b *Buffer) Get(x, y int) (*Cell, bool) {
	return b.Data.GetRef(x, y)
}

func (b *Buffer) Set(x int, y int, v byte, s tcell.Style) {
	b.Data.Set(x, y, Cell{
		Value: v,
		Style: s,
	})
}

func (b *Buffer) SetString(x int, y int, s []byte, st tcell.Style) {
	for i, ch := range s {
		b.Data.Set(x+i, y, Cell{
			Value: ch,
			Style: st,
		})
	}
}

func (b *Buffer) Import(r io.Reader) error {
	var lines [][]byte
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		lines = append(lines, bytes.Clone(sc.Bytes()))
		if err := sc.Err(); err != nil {
			return err
		}
	}

	if len(lines) == 0 || len(lines[0]) == 0 {
		return errors.New("Empty data")
	}

	b.Data = MakeGrid(len(lines[0]), len(lines), Cell{})

	for y := range b.Data.Height {
		for x := range b.Data.Width {
			c, _ := b.Data.GetRef(x, y)
			c.Value = lines[y][x]
		}
	}

	return nil
}

func (b *Buffer) Export(w io.Writer) error {
	for y := range b.Data.Height {
		for x := range b.Data.Width {
			if _, err := fmt.Fprintf(w, "%c", b.Data.MustGet(x, y).Value); err != nil {
				return err
			}
		}
		if y != b.Data.Height-1 {
			if _, err := fmt.Fprintf(w, "%c", '\n'); err != nil {
				return err
			}
		}
	}
	return nil
}

func (b *Buffer) Load(r io.Reader) error {
	var magic int64

	if err := binary.Read(r, binary.BigEndian, &magic); err != nil {
		return err
	}
	if magic != magicNumber {
		return errors.New("Invalid magic number")
	}

	var width, height int32

	if err := binary.Read(r, binary.BigEndian, &width); err != nil {
		return err
	}

	if err := binary.Read(r, binary.BigEndian, &height); err != nil {
		return err
	}

	if width <= 0 || height <= 0 {
		return errors.New("Width and height must be positive")
	}

	b.Data = MakeGrid(int(width), int(height), Cell{})
	for y := range int(height) {
		for x := range int(width) {
			rf, _ := b.Data.GetRef(x, y)
			var rawStyle [3]uint64
			if err := binary.Read(r, binary.BigEndian, &rf.Value); err != nil {
				return err
			}
			if err := binary.Read(r, binary.BigEndian, &rawStyle); err != nil {
				return err
			}
			rf.Style = tcell.StyleDefault.
				Foreground(tcell.Color(rawStyle[0])).
				Background(tcell.Color(rawStyle[1])).
				Attributes(tcell.AttrMask(rawStyle[2]))
		}
	}
	return nil
}

func (b *Buffer) Save(w io.Writer) error {
	if err := binary.Write(w, binary.BigEndian, magicNumber); err != nil {
		return err
	}

	width, height := int32(b.Data.Width), int32(b.Data.Height)

	if err := binary.Write(w, binary.BigEndian, width); err != nil {
		return err
	}

	if err := binary.Write(w, binary.BigEndian, height); err != nil {
		return err
	}

	for y := range b.Data.Height {
		for x := range b.Data.Width {
			c, _ := b.Data.GetRef(x, y)
			fg, bg, attr := c.Style.Decompose()
			vals := [3]uint64{uint64(fg), uint64(bg), uint64(attr)}
			if err := binary.Write(w, binary.BigEndian, c.Value); err != nil {
				return err
			}
			if err := binary.Write(w, binary.BigEndian, vals); err != nil {
				return err
			}
		}
	}

	return nil
}
