package game

import (
	"math/rand/v2"
)

// growStraightChance10 is P(straight)/10 when both straight and turn tail steps exist
// during grow algorithm extensions in tryGrowPartition.
const growStraightChance10 = 9

// tryGrowPartition builds initial seed paths then repeatedly extends tails until no move remains.
func tryGrowPartition(w, h, nHeads int, rng *rand.Rand) ([][]Point, bool) {
	paths, ok := seedGrowPaths(w, h, nHeads, rng)
	if !ok {
		return nil, false
	}

	occupied := make([]bool, w*h)
	for _, path := range paths {
		for _, p := range path {
			occupied[p.Y*w+p.X] = true
		}
	}

	for {
		extended := false
		order := rng.Perm(len(paths))
		for _, pi := range order {
			path := paths[pi]
			if len(path) < 2 {
				return nil, false
			}
			tail := path[len(path)-1]
			prev := path[len(path)-2]
			pathSet := make(map[Point]struct{}, len(path)+1)
			for _, p := range path {
				pathSet[p] = struct{}{}
			}
			cands := neighborPoints(tail, prev, w, h, occupied, pathSet)
			hx, hy := path[0].X, path[0].Y
			fire := oppositeDir(directionFromTo(hx, hy, path[1].X, path[1].Y))
			write := 0
			for _, c := range cands {
				if !cellOnOpenRayFromHead(hx, hy, fire, c.X, c.Y, w, h) {
					cands[write] = c
					write++
				}
			}
			cands = cands[:write]
			if len(cands) == 0 {
				continue
			}
			next := pickBiasedTailStep(prev, tail, cands, rng, growStraightChance10)
			path = append(path, next)
			paths[pi] = path
			occupied[next.Y*w+next.X] = true
			extended = true
		}
		if !extended {
			break
		}
	}

	return paths, true
}

// seedGrowPaths places nHeads disjoint two-cell paths (head + one body) with random orientation.
func seedGrowPaths(w, h, nHeads int, rng *rand.Rand) ([][]Point, bool) {
	wh := w * h
	occupied := make([]bool, wh)
	var heads []Point
	var paths [][]Point

	maxSeedAttempts := 400 * nHeads
	if maxSeedAttempts < 800 {
		maxSeedAttempts = 800
	}

	for k := 0; k < nHeads; k++ {
		placed := false
		for attempt := 0; attempt < maxSeedAttempts; attempt++ {
			hx := rng.IntN(w)
			hy := rng.IntN(h)
			if occupied[hy*w+hx] {
				continue
			}
			fire := Direction(rng.IntN(4))
			if rayHitsPreviousHead(hx, hy, fire, heads, w, h) {
				continue
			}
			bdx, bdy := Delta(oppositeDir(fire))
			bx, by := hx+bdx, hy+bdy
			if bx < 0 || bx >= w || by < 0 || by >= h {
				continue
			}
			if occupied[by*w+bx] {
				continue
			}
			path := []Point{{hx, hy}, {bx, by}}
			occupied[hy*w+hx] = true
			occupied[by*w+bx] = true
			heads = append(heads, Point{hx, hy})
			paths = append(paths, path)
			placed = true
			break
		}
		if !placed {
			return nil, false
		}
	}
	return paths, true
}

// rayHitsPreviousHead reports whether the open ray from head (hx,hy) in fire direction passes any prior head.
func rayHitsPreviousHead(hx, hy int, fire Direction, heads []Point, w, h int) bool {
	dx, dy := Delta(fire)
	for cx, cy := hx+dx, hy+dy; cx >= 0 && cx < w && cy >= 0 && cy < h; cx, cy = cx+dx, cy+dy {
		for _, hp := range heads {
			if hp.X == cx && hp.Y == cy {
				return true
			}
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
