package game

import (
	"errors"
)

// boardFromPaths rasterizes polylines into a Board via paintPath (each path must be disjoint).
// The production generator builds boards directly from sc.glyphAt; this helper is
// kept for tests and standalone callers that have a [][]Point and want a Board.
func boardFromPaths(paths [][]Point, w, h int) (Board, error) {
	grid := make([]rune, w*h)
	for _, path := range paths {
		if err := paintPath(grid, w, path); err != nil {
			return Board{}, err
		}
	}
	b := NewBoard(w, h)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			b.Set(x, y, Cell{R: grid[y*w+x]})
		}
	}
	return b, nil
}

// paintPath writes one polyline into grid: head at path[0], tail at path[len-1], using wire corners.
func paintPath(grid []rune, w int, path []Point) error {
	if len(path) < 2 {
		return errors.New("path too short")
	}
	hx, hy := path[0].X, path[0].Y
	i0 := hy*w + hx
	if grid[i0] != 0 {
		return errors.New("cell occupied")
	}
	dBody := directionFromTo(hx, hy, path[1].X, path[1].Y)
	grid[i0] = headRuneForFire(oppositeDir(dBody))

	for i := 1; i < len(path); i++ {
		px, py := path[i].X, path[i].Y
		idx := py*w + px
		if grid[idx] != 0 {
			return errors.New("cell occupied")
		}
		dPrev := directionFromTo(px, py, path[i-1].X, path[i-1].Y)
		if i < len(path)-1 {
			dNext := directionFromTo(px, py, path[i+1].X, path[i+1].Y)
			grid[idx] = wireRuneTwo(dPrev, dNext)
		} else {
			grid[idx] = wireRuneOne(dPrev)
		}
	}
	return nil
}

// headRuneForFire returns the Unicode head rune that fires in the given direction.
func headRuneForFire(fire Direction) rune {
	switch fire {
	case North:
		return '▲'
	case South:
		return '▼'
	case East:
		return '▶'
	case West:
		return '◀'
	default:
		return '▲'
	}
}

// wireRuneOne returns the wire rune for a degree-1 (tail) cell opening toward dPrev (into the body).
func wireRuneOne(d Direction) rune {
	switch d {
	case North, South:
		return '│'
	case East, West:
		return '─'
	default:
		return '│'
	}
}

// wireRuneTwo returns the corner or straight wire for an internal cell with neighbors along a and b.
func wireRuneTwo(a, b Direction) rune {
	if a == oppositeDir(b) {
		if a == North || a == South {
			return '│'
		}
		return '─'
	}
	set := map[Direction]bool{a: true, b: true}
	switch {
	case set[North] && set[East]:
		return '└'
	case set[North] && set[West]:
		return '┘'
	case set[South] && set[East]:
		return '┌'
	case set[South] && set[West]:
		return '┐'
	default:
		return '│'
	}
}
