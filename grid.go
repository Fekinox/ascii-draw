package main

import (
	"bufio"
	"fmt"
	"io"
	"unicode"
)

type Grid[T any] struct {
	data   []T
	Width  int
	Height int
}

type Position struct {
	X int
	Y int
}

func MakeGrid[T any](width, height int, def T) Grid[T] {
	data := make([]T, width*height)
	for i := 0; i < width*height; i++ {
		data[i] = def
	}

	return Grid[T]{
		data:   data,
		Width:  width,
		Height: height,
	}
}

func MakeGridWith[T any](width, height int, gen func(x, y int) T) Grid[T] {
	data := make([]T, width*height)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			data[y*width+x] = gen(x, y)
		}
	}

	return Grid[T]{
		data:   data,
		Width:  width,
		Height: height,
	}
}

func GridFromSlice[T any](width, height int, elems ...T) Grid[T] {
	data := make([]T, width*height)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			data[y*width+x] = elems[y*width+x]
		}
	}

	return Grid[T]{
		data:   data,
		Width:  width,
		Height: height,
	}
}

func GridFromSlices[T any](slices ...[]T) Grid[T] {
	height := len(slices)
	width := len(slices[0])
	data := make([]T, width*height)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			data[y*width+x] = slices[y][x]
		}
	}

	return Grid[T]{
		data:   data,
		Width:  width,
		Height: height,
	}
}

// Ignores non-Unicode characters.
func GridFromReader(r io.Reader) (Grid[byte], error) {
	var lines [][]byte
	var maxWidth = 0
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		var l []byte
		for _, r := range sc.Text() {
			if r > unicode.MaxASCII {
				l = append(l, '?')
			} else if r == '\t' {
				l = append(l, ' ', ' ', ' ', ' ')
			} else if unicode.IsSpace(r) {
				l = append(l, ' ')
			} else if r < 0x20 {
				l = append(l, '?')
			} else {
				l = append(l, byte(r))
			}
		}
		maxWidth = max(maxWidth, len(l))
		lines = append(lines, l)
	}

	if sc.Err() != nil {
		return Grid[byte]{}, sc.Err()
	}

	res := MakeGrid(maxWidth, len(lines), byte(' '))
	for y, ln := range lines {
		for x, c := range ln {
			res.Set(x, y, c)
		}
	}

	return res, nil
}

func (g *Grid[T]) InBounds(x, y int) bool {
	return x >= 0 && x < g.Width && y >= 0 && y < g.Height
}

func (g *Grid[T]) Get(x int, y int) (T, bool) {
	if !g.InBounds(x, y) {
		return *new(T), false
	}

	return g.data[y*g.Width+x], true
}

func (g *Grid[T]) MustGet(x, y int) T {
	if !g.InBounds(x, y) {
		panic(
			fmt.Sprintf(
				"Out of bounds index %d %d for grid of size %d %d",
				x, y,
				g.Width, g.Height,
			),
		)

	}
	return g.data[y*g.Width+x]
}

func (g *Grid[T]) Set(x, y int, val T) bool {
	if !g.InBounds(x, y) {
		return false
	}

	g.data[y*g.Width+x] = val

	return true
}

func (g *Grid[T]) GetRef(x, y int) (*T, bool) {
	if !g.InBounds(x, y) {
		return nil, false
	}

	return &g.data[y*g.Width+x], true
}

func (g *Grid[T]) Resize(ox, oy int, neww, newh int, def T) Grid[T] {
	return MakeGridWith(neww, newh, func(x, y int) T {
		xx := x - ox
		yy := y - oy
		val, ok := g.Get(xx, yy)
		if ok {
			return val
		} else {
			return def
		}
	})
}

func (g *Grid[T]) ShallowClone() Grid[T] {
	return MakeGridWith(g.Width, g.Height, func(x, y int) T {
		return g.MustGet(x, y)
	})
}
