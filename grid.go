package main

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

func GridFromStrings(strings ...string) Grid[rune] {
	slices := make([][]rune, len(strings))
	for i := 0; i < len(strings); i++ {
		slices[i] = []rune(strings[i])
	}

	return GridFromSlices(slices...)
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
		panic("Out of bounds")
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

func ShiftedDifference(g Grid[bool], dx, dy int) Grid[bool] {
	return MakeGridWith(g.Width, g.Height, func(x, y int) bool {
		valAtOldPos, outOfBoundsAtOldPos := g.Get(x-dx, y-dy)
		return g.MustGet(x, y) && (!valAtOldPos || outOfBoundsAtOldPos)
	})
}
