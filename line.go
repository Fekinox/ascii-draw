package main

func linePositionsLow(ax, ay, bx, by int) (pos []Position) {
	dx, dy := bx-ax, by-ay
	yi := 1
	if dy < 0 {
		yi = -1
		dy = -dy
	}
	D := 2*dy - dx
	y := ay

	for x := ax; x <= bx; x++ {
		pos = append(pos, Position{X: x, Y: y})
		if D > 0 {
			y += yi
			D += 2 * (dy - dx)
		} else {
			D += 2 * dy
		}
	}
	return
}

func linePositionsHigh(ax, ay, bx, by int) (pos []Position) {
	dx, dy := bx-ax, by-ay
	xi := 1
	if dx < 0 {
		xi = -1
		dx = -dx
	}
	D := 2*dx - dy
	x := ax

	for y := ay; y <= by; y++ {
		pos = append(pos, Position{X: x, Y: y})
		if D > 0 {
			x += xi
			D += 2 * (dx - dy)
		} else {
			D += 2 * dx
		}
	}
	return
}

// Given two points, computes all of the pixel points between them using Bresenham's line algorithm.
func LinePositions(ax, ay, bx, by int) []Position {
	dx, dy := bx-ax, by-ay
	if max(dy, -dy) < max(dx, -dx) {
		if dx < 0 {
			ax, ay, bx, by = bx, by, ax, ay
		}
		return linePositionsLow(ax, ay, bx, by)
	} else {
		if dy < 0 {
			ax, ay, bx, by = bx, by, ax, ay
		}
		return linePositionsHigh(ax, ay, bx, by)
	}
}
