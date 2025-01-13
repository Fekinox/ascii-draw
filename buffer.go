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

type Cell struct {
	Value byte
	Style tcell.Style
}

func Encode(c *Cell) uint16 {
	choppedValue := uint16(c.Value & 127)
	fg, bg, _ := c.Style.Decompose()
	var fgc, bgc uint16
	if fg == tcell.ColorDefault {
		fgc = 16
	} else {
		fgc = uint16(fg - tcell.ColorValid)
	}
	if bg == tcell.ColorDefault {
		bgc = 16
	} else {
		bgc = uint16(bg - tcell.ColorValid)
	}
	return choppedValue + ((fgc + bgc*17) << 7)
}

func Decode(u uint16, c *Cell) {
	c.Value = byte(u & 127)
	u = u >> 7
	fgc, bgc := u%17, u/17
	var fg, bg tcell.Color
	if fgc != 16 {
		fg = tcell.Color(fgc) + tcell.ColorValid
	}
	if bgc != 16 {
		bg = tcell.Color(bgc) + tcell.ColorValid
	}
	c.Style = c.Style.Foreground(fg).Background(bg)
}

type Buffer struct {
	Data            Grid[Cell]
	activeSelection bool
	SelectionMask   Grid[bool]
}

func MakeBuffer(width, height int) *Buffer {
	b := &Buffer{
		Data:          MakeGrid(width, height, Cell{Value: ' '}),
		SelectionMask: MakeGrid(width, height, false),
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
	// For safety's sake, turn every non-7-bit graphic ASCII character into a space.
	// SetContent already does this, but I wanna be extra safe
	if v <= 0x20 || v >= 0x7f {
		v = ' '
	}
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
			var u uint16
			if err := binary.Read(r, binary.BigEndian, &u); err != nil {
				return err
			}
			rf, _ := b.Data.GetRef(x, y)
			Decode(u, rf)
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
			if err := binary.Write(w, binary.BigEndian, Encode(c)); err != nil {
				return err
			}
		}
	}

	return nil
}

func (b *Buffer) Clone() *Buffer {
	return &Buffer{
		Data: b.Data.ShallowClone(),
	}
}

func (b *Buffer) Clear() {
	for y := range b.Data.Height {
		for x := range b.Data.Width {
			b.Data.Set(x, y, Cell{Value: ' '})
		}
	}
}

func (b *Buffer) Deselect() {
	b.activeSelection = false
	for y := range b.Data.Height {
		for x := range b.Data.Width {
			b.SelectionMask.Set(x, y, false)
		}
	}
}

func (b *Buffer) Translate(
	other *Buffer,
	mask Grid[bool],
	topLeft Position,
	dx, dy int,
) {
	b.Clear()

	for y := range b.Data.Height {
		for x := range b.Data.Width {
			var v Cell
			if m, ok := mask.Get(x-dx-topLeft.X, y-dy-topLeft.Y); m && ok {
				// If (x,y) - (dx, dy) + (tlx, tly) is in the mask, return the value of
				// (x,y) - (dx,dy).
				if val, ok := other.Data.Get(x-dx, y-dy); ok {
					v = val
				}
			} else if m, ok := mask.Get(x-topLeft.X, y-topLeft.Y); !m || !ok {
				// If (x,y) + (tlx,tly) is not in the mask, return the value of (x,y).
				v = other.Data.MustGet(x, y)
			}
			b.Data.Set(x, y, v)
		}
	}
}

func (b *Buffer) TranslateBlankTransparent(
	other *Buffer,
	mask Grid[bool],
	topLeft Position,
	dx, dy int,
) {
	b.Clear()

	for y := range b.Data.Height {
		for x := range b.Data.Width {
			if m, ok := mask.Get(x-dx-topLeft.X, y-dy-topLeft.Y); m && ok {
				// If (x,y) - (dx, dy) + (tlx, tly) is in the mask, return the value of
				// (x,y) - (dx,dy).
				val, ok := other.Data.Get(x-dx, y-dy)
				if ok && val.Value != ' ' {
					b.Data.Set(x, y, val)
					continue
				}
			}
			if m, ok := mask.Get(x-topLeft.X, y-topLeft.Y); !m || !ok {
				b.Data.Set(x, y, other.Data.MustGet(x, y))
			}
		}
	}
}

func (b *Buffer) FillRegion(x, y, w, h int, cell Cell) {
	minX := max(0, min(b.Data.Width, x))
	minY := max(0, min(b.Data.Height, y))
	maxX := max(0, min(b.Data.Width, x+w))
	maxY := max(0, min(b.Data.Height, y+h))
	for yy := minY; yy < maxY; yy++ {
		for xx := minX; xx < maxX; xx++ {
			b.Data.Set(xx, yy, cell)
		}
	}
}

func (b *Buffer) SetSelection(mask Grid[bool], topLeft Position) {
	b.Deselect()
	for y := range b.Data.Height {
		for x := range b.Data.Width {
			if inMask, ok := mask.Get(x-topLeft.X, y-topLeft.Y); ok && inMask {
				b.SelectionMask.Set(x, y, true)
			}
		}
	}
	b.activeSelection = true
}

func (b *Buffer) BrushStrokes(radius int, cell Cell, points []Position) {
	for _, pt := range points {
		b.FillRegion(pt.X-radius/2, pt.Y-radius/2, radius, radius, cell)
	}
}

func (b *Buffer) Stamp(other *Buffer, clipboard Grid[Cell], points []Position) {
	dx, dy := -clipboard.Width/2, -clipboard.Height/2
	for _, pt := range points {
		for y := range clipboard.Height {
			for x := range clipboard.Width {
				c := clipboard.MustGet(x, y)
				if c.Value != ' ' {
					b.Data.Set(x+pt.X+dx, y+pt.Y+dy, c)
				}
			}
		}
	}
}

func (b *Buffer) ClearSelection() {
	for y := range b.Data.Height {
		for x := range b.Data.Width {
			if b.SelectionMask.MustGet(x, y) {
				b.Data.Set(x, y, Cell{Value: ' '})
			}
		}
	}
}

func (b *Buffer) CopySelection() Grid[Cell] {
	minX, maxX := b.Data.Width-1, 0
	minY, maxY := b.Data.Height-1, 0

	for y := range b.Data.Height {
		for x := range b.Data.Width {
			if b.SelectionMask.MustGet(x, y) {
				minX, maxX = min(minX, x), max(maxX, x)
				minY, maxY = min(minY, y), max(maxY, y)
			}
		}
	}

	res := MakeGrid(maxX-minX+1, maxY-minY+1, Cell{Value: ' '})
	for y := range res.Height {
		for x := range res.Width {
			if b.SelectionMask.MustGet(x+minX, y+minY) {
				res.Set(x, y, b.Data.MustGet(x+minX, y+minY))
			}
		}
	}

	return res
}
