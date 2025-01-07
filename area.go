package main

type Area struct {
	X      int
	Y      int
	Width  int
	Height int
}

func (a Area) Top() int    { return a.Y }
func (a Area) Bottom() int { return a.Y + a.Height }
func (a Area) Left() int   { return a.X }
func (a Area) Right() int  { return a.X + a.Width }

func (a Area) Intersects(other Area) bool {
	return !((a.Right() < other.Left() || other.Right() < a.Left()) ||
		(a.Bottom() < other.Top() || other.Bottom() < a.Top()))
}

func (a Area) Intersection(other Area) (Area, bool) {
	if !a.Intersects(other) {
		return Area{}, false
	}

	top := max(a.Top(), other.Top())
	bottom := min(a.Bottom(), other.Bottom())
	left := max(a.Left(), other.Left())
	right := min(a.Right(), other.Right())

	return Area{
		X:      left,
		Y:      top,
		Width:  right - left,
		Height: bottom - top,
	}, true
}

func (a Area) Contains(x, y int) bool {
	nx := x - a.X
	ny := y - a.Y
	return nx >= 0 && nx < a.Width && ny >= 0 && ny < a.Height
}

func (a Area) Inset(width, height int) Area {
	return Area{
		X:      a.X + (a.Width-width)/2,
		Y:      a.Y + (a.Height-height)/2,
		Width:  width,
		Height: height,
	}
}
