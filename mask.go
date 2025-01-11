package main

import (
	"math"
	"slices"
)

func CreateMask(positions []Position) (Position, Grid[bool]) {
	if len(positions) == 0 {
		return Position{}, MakeGrid(0, 0, false)
	}

	tlc, brc := positions[0], positions[0]
	for _, p := range positions[1:] {
		tlc.X = min(tlc.X, p.X)
		tlc.Y = min(tlc.Y, p.Y)
		brc.X = max(brc.X, p.X)
		brc.Y = max(brc.Y, p.Y)
	}

	width, height := brc.X-tlc.X+1, brc.Y-tlc.Y+1
	mask := MakeGrid(width, height, false)

	// Fill in points on the line just for consistency's sake
	j := len(positions) - 1
	for i, p1 := range positions {
		p2 := positions[j]
		linePoints := LinePositions(p1.X, p1.Y, p2.X, p2.Y)
		for _, p := range linePoints {
			mask.Set(p.X-tlc.X, p.Y-tlc.Y, true)
		}
		j = i
	}

	var nodes int
	nodeX := make([]int, len(positions))
	for yy := range mask.Height {
		nodes = 0
		j := len(positions) - 1
		for i := 0; i < len(positions); i++ {
			p1, p2 := positions[i], positions[j]
			p1x, p1y, p2x, p2y := p1.X-tlc.X, p1.Y-tlc.Y, p2.X-tlc.X, p2.Y-tlc.Y
			if p1y < yy && p2y >= yy || p2y < yy && p1y >= yy {
				t := (float64(yy) - float64(p1y)) / (float64(p2y) - float64(p1y))
				nx := math.Round(float64(p1x) + t*float64(p2x-p1x))
				nodeX[nodes] = int(nx)
				nodes++
			}
			j = i
		}
		slices.Sort(nodeX[:nodes])
		for i := 0; i < nodes; i += 2 {
			if nodeX[i] >= mask.Width-1 {
				break
			}
			if nodeX[i+1] > 0 {
				if nodeX[i] < 0 {
					nodeX[i] = 0
				}
				if nodeX[i+1] > mask.Width-1 {
					nodeX[i+1] = mask.Width - 1
				}
				for xx := nodeX[i]; xx < nodeX[i+1]; xx++ {
					mask.Set(xx, yy, true)
				}
			}
		}
	}

	return tlc, mask
}
