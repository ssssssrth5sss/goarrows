package game

import (
	"errors"
	"math/rand/v2"
)

// GenGrow is the only supported procedural generation algorithm.
const GenGrow = "grow"

// growStraightChance10 is P(straight)/10 when both straight and turn tail steps exist
// during grow algorithm extensions in tryGrowPartition.
const growStraightChance10 = 9

// GenerateBoard fills a w×h grid with the grow procedural algorithm.
func GenerateBoard(w, h int, rng *rand.Rand) (Board, error) {
	return generateFullBoardGrow(w, h, rng)
}

// cellOnOpenRayFromHead reports whether (px, py) lies on the open ray from (hx, hy) in
// direction fire: the first cell is (hx, hy)+Delta(fire), excluding the head cell itself.
// Matches RayEscapes ray traversal.
func cellOnOpenRayFromHead(hx, hy int, fire Direction, px, py, w, h int) bool {
	dx, dy := Delta(fire)
	for cx, cy := hx+dx, hy+dy; cx >= 0 && cx < w && cy >= 0 && cy < h; cx, cy = cx+dx, cy+dy {
		if cx == px && cy == py {
			return true
		}
	}
	return false
}

// neighborPoints returns empty orthogonal steps from tail toward extending the polyline:
// in bounds, not backtracking to prev, not in occupied, and not on the current path.
func neighborPoints(tail, prev Point, w, h int, occupied []bool, pathSet map[Point]struct{}) []Point {
	var out []Point
	for _, d := range []Direction{North, East, South, West} {
		dx, dy := Delta(d)
		nx, ny := tail.X+dx, tail.Y+dy
		if nx < 0 || nx >= w || ny < 0 || ny >= h {
			continue
		}
		np := Point{nx, ny}
		if np == prev {
			continue
		}
		if occupied[ny*w+nx] {
			continue
		}
		if _, ok := pathSet[np]; ok {
			continue
		}
		out = append(out, np)
	}
	return out
}

// pickBiasedTailStep chooses the next cell when extending a polyline tail. When both a straight
// continuation and a turn are legal, straightChance10 out of 10 rolls pick straight.
func pickBiasedTailStep(prev, tail Point, cands []Point, rng *rand.Rand, straightChance10 int) Point {
	if len(cands) == 1 {
		return cands[0]
	}
	incoming := directionFromTo(prev.X, prev.Y, tail.X, tail.Y)
	var straight, turn []Point
	for _, c := range cands {
		out := directionFromTo(tail.X, tail.Y, c.X, c.Y)
		if out == incoming {
			straight = append(straight, c)
		} else {
			turn = append(turn, c)
		}
	}
	if len(turn) > 0 && len(straight) > 0 {
		if rng.IntN(10) < straightChance10 {
			return straight[rng.IntN(len(straight))]
		}
		return turn[rng.IntN(len(turn))]
	}
	return cands[rng.IntN(len(cands))]
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

