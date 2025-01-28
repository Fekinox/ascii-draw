package main

type Position struct {
	X int
	Y int
}

func SquaredDistance(x1, y1, x2, y2 int) int {
	return (x1-x2)*(x1-x2) + (y1-y2)*(y1-y2)
}
