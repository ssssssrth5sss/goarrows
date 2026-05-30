package game

import (
	"math/rand/v2"
)

// growStraightChance10 is P(straight)/10 when both straight and turn tail steps exist
// during grow algorithm extensions in tryGrowPartitionBuf.
const growStraightChance10 = 9

// tryGrowPartitionBuf builds initial seed paths then repeatedly extends tails until no
// move remains, reusing scratch buffers across attempts. It maintains sc.glyphAt
// incrementally so the caller can build the final Board directly without a second
// rasterization pass through boardFromPathsBuf.
func tryGrowPartitionBuf(w, h, nHeads int, rng *rand.Rand, sc *genScratch) ([][]Point, bool) {
	paths, ok := seedGrowPathsBuf(w, h, nHeads, rng, sc)
	if !ok {
		return nil, false
	}

	wh := w * h
	occupied := sc.occupied[:wh]
	glyphAt := sc.glyphAt[:wh]
	clear(occupied)
	// seedGrowPathsBuf already populated glyphAt for the seed cells; mirror that
	// into the extension-phase occupied bitmap.
	for _, path := range paths {
		for _, p := range path {
			occupied[p.Y*w+p.X] = true
		}
	}

	// fireMasks[pi][y*w+x] is true iff (x,y) lies on path pi's open firing ray.
	// Head and body direction are fixed at seed time, so the ray cells never change.
	maskBuf := sc.fireMaskBuf[:len(paths)*wh]
	clear(maskBuf)
	fireMasks := make([][]bool, len(paths))
	for pi, path := range paths {
		fireMasks[pi] = maskBuf[pi*wh : (pi+1)*wh]
		hx, hy := path[0].X, path[0].Y
		fire := oppositeDir(directionFromTo(hx, hy, path[1].X, path[1].Y))
		dx, dy := Delta(fire)
		for cx, cy := hx+dx, hy+dy; cx >= 0 && cx < w && cy >= 0 && cy < h; cx, cy = cx+dx, cy+dy {
			fireMasks[pi][cy*w+cx] = true
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
			cands := neighborPoints(tail, prev, w, h, occupied)
			fm := fireMasks[pi]
			write := 0
			for _, c := range cands {
				if !fm[c.Y*w+c.X] {
					cands[write] = c
					write++
				}
			}
			cands = cands[:write]
			if len(cands) == 0 {
				continue
			}
			next := pickBiasedTailStep(prev, tail, cands, rng, growStraightChance10)
			// Tail transitions from degree-1 wire to degree-2 internal cell; the new
			// tail becomes a degree-1 wire. Update glyphAt so the caller can build the
			// final Board without a second rasterization pass. Note: glyphs are not
			// checked for accidental cross-path port links here — that filter caused
			// solver-rejection rates to spike (dense link-free growth tends to produce
			// unsolvable boards). ValidatePartialBoard runs after grow and rejects the
			// rare accidental links that do occur.
			glyphAt[tail.Y*w+tail.X] = wireRuneTwo(
				directionFromTo(tail.X, tail.Y, prev.X, prev.Y),
				directionFromTo(tail.X, tail.Y, next.X, next.Y),
			)
			glyphAt[next.Y*w+next.X] = wireRuneOne(directionFromTo(next.X, next.Y, tail.X, tail.Y))
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


// seedGrowPathsBuf places nHeads disjoint two-cell paths (head + one body) with
// random orientation, reusing scratch buffers. A head bitmap turns the
// previously-O(K) "ray hits another head?" check into O(1) per cell. As paths
// are placed sc.glyphAt is populated so the extension phase and the final Board
// build can read glyphs directly without re-rasterizing the paths.
func seedGrowPathsBuf(w, h, nHeads int, rng *rand.Rand, sc *genScratch) ([][]Point, bool) {
	wh := w * h
	occupied := sc.seedOcc[:wh]
	headBM := sc.seedHeadBM[:wh]
	glyphAt := sc.glyphAt[:wh]
	clear(occupied)
	clear(headBM)
	clear(glyphAt)
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
			if rayHitsHeadBitmap(hx, hy, fire, headBM, w, h) {
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
			headBM[hy*w+hx] = true
			glyphAt[hy*w+hx] = headRuneForFire(fire)
			glyphAt[by*w+bx] = wireRuneOne(directionFromTo(bx, by, hx, hy))
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

// rayHitsHeadBitmap reports whether the open ray from (hx,hy) in fire direction
// crosses a previously placed head. headBM[y*w+x] is true at every head cell.
func rayHitsHeadBitmap(hx, hy int, fire Direction, headBM []bool, w, h int) bool {
	dx, dy := Delta(fire)
	for cx, cy := hx+dx, hy+dy; cx >= 0 && cx < w && cy >= 0 && cy < h; cx, cy = cx+dx, cy+dy {
		if headBM[cy*w+cx] {
			return true
		}
	}
	return false
}

// neighborPoints returns empty orthogonal steps from tail toward extending the polyline:
// in bounds, not backtracking to prev, and not in occupied (which already contains
// every cell of every path, so an explicit path-set check is redundant).
func neighborPoints(tail, prev Point, w, h int, occupied []bool) []Point {
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
