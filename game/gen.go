package game

import (
	"fmt"
	"math/rand/v2"
)

// GenGrow is the only supported procedural generation algorithm.
const GenGrow = "grow"

// GenerateBoard fills a w×h grid with the grow procedural algorithm.
func GenerateBoard(w, h int, rng *rand.Rand) (Board, error) {
	return generateFullBoardGrow(w, h, rng)
}

// targetArrowCountForSide returns how many arrow polylines to use for an N×N layer:
// N < 6 → N; N < 10 → N*N/6; otherwise → N*N/10 (integer division).
func targetArrowCountForSide(n int) int {
	switch {
	case n < 6:
		return n
	case n < 10:
		return n * n / 6
	default:
		return n * n / 10
	}
}

// clampArrowCount caps the target polyline count to [1, wh/2] so seeds fit on the grid.
func clampArrowCount(targetArrows, wh int) int {
	maxArrows := wh / 2
	if maxArrows < 1 {
		maxArrows = 1
	}
	if targetArrows > maxArrows {
		targetArrows = maxArrows
	}
	if targetArrows < 1 {
		targetArrows = 1
	}
	return targetArrows
}

// generateFullBoardGrow seeds arrow heads with a one-cell body (count from targetArrowCountForSide),
// extends tails at random until stuck, then accepts only if ValidatePartialBoard,
// growPlayfulEnough (at most half the heads have a clear shot at start), and
// VerifyGreedyFirstClearsBoard succeed.
func generateFullBoardGrow(w, h int, rng *rand.Rand) (Board, error) {
	if w <= 0 || h <= 0 {
		return Board{}, fmt.Errorf("gen: invalid size %d×%d", w, h)
	}
	wh := w * h
	if wh < 2 {
		return Board{}, fmt.Errorf("gen: need at least 2 cells (got %d×%d)", w, h)
	}

	n := min(w, h)
	nHeads := clampArrowCount(targetArrowCountForSide(n), wh)
	maxTries := 8000 + 100*wh
	if maxTries > 60000 {
		maxTries = 60000
	}
	for attempt := 0; attempt < maxTries; attempt++ {
		r := rand.New(rand.NewPCG(rng.Uint64(), rng.Uint64()))
		paths, ok := tryGrowPartition(w, h, nHeads, r)
		if !ok {
			continue
		}
		b, err := boardFromPaths(paths, w, h)
		if err != nil {
			continue
		}
		if err := ValidatePartialBoard(b); err != nil {
			continue
		}
		if !growPlayfulEnough(b) {
			continue
		}
		if !VerifyGreedyFirstClearsBoard(b) {
			continue
		}
		return b, nil
	}
	return Board{}, fmt.Errorf("gen: could not build grow board for %d×%d", w, h)
}

// growPlayfulEnough rejects boards that are too easy at the start: at most half the heads
// may have a clear firing ray (RayEscapes). For a single head the check is skipped so
// solvable tiny cases are still possible.
func growPlayfulEnough(b Board) bool {
	total := 0
	fireable := 0
	for y := 0; y < b.H; y++ {
		for x := 0; x < b.W; x++ {
			if !b.At(x, y).IsHead() {
				continue
			}
			total++
			if RayEscapes(b, x, y) {
				fireable++
			}
		}
	}
	if total <= 1 {
		return true
	}
	return 2*fireable <= total
}
