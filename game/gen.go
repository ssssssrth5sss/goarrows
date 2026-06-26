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

// genScratch holds buffers reused across generation attempts to avoid GC pressure.
// All slices are sized for the worst case (nHeads heads on a w×h grid) and reset
// via clear() at the start of each attempt.
type genScratch struct {
	occupied    []bool  // size w*h: tryGrowPartition path-occupancy bitmap
	glyphAt     []rune  // size w*h: tracks rune at each occupied cell during grow
	seedOcc     []bool  // size w*h: seedGrowPaths occupancy bitmap
	seedHeadBM  []bool  // size w*h: seedGrowPaths head bitmap (replaces O(K) scan)
	fireMaskBuf []bool  // size nHeads*w*h: backing storage for per-path firing-ray masks
	verifyCells []Cell  // size w*h: verifySolvableFastBuf scratch board
	rayClear    []bool  // size nHeads
	alive       []bool  // size nHeads
	queue       []int   // capacity nHeads
	rayCellsBuf [][]int // [nHeads][]int, reused per call
	cellHeadsBy [][]int // [w*h][]int reverse index
}

// newGenScratch allocates the per-call scratch buffers used by generateFullBoardGrow.
func newGenScratch(w, h, nHeads int) *genScratch {
	wh := w * h
	return &genScratch{
		occupied:    make([]bool, wh),
		glyphAt:     make([]rune, wh),
		seedOcc:     make([]bool, wh),
		seedHeadBM:  make([]bool, wh),
		fireMaskBuf: make([]bool, nHeads*wh),
		verifyCells: make([]Cell, wh),
		rayClear:    make([]bool, nHeads),
		alive:       make([]bool, nHeads),
		queue:       make([]int, 0, nHeads),
		rayCellsBuf: make([][]int, nHeads),
		cellHeadsBy: make([][]int, wh),
	}
}

// generateFullBoardGrow seeds arrow heads with a one-cell body (count from targetArrowCountForSide),
// extends tails at random until stuck, then accepts only if ValidatePartialBoard,
// growPlayfulEnoughHeads (at most half the heads have a clear shot at start), and
// VerifySolvableFast succeed.
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
	scratch := newGenScratch(w, h, nHeads)
	heads := make([]Point, 0, nHeads)
	for attempt := 0; attempt < maxTries; attempt++ {
		r := rand.New(rand.NewPCG(rng.Uint64(), rng.Uint64()))
		paths, ok := tryGrowPartitionBuf(w, h, nHeads, r, scratch)
		if !ok {
			continue
		}
		// tryGrowPartitionBuf populated scratch.glyphAt incrementally — build the
		// Board directly from it without a second rasterization pass through
		// boardFromPathsBuf.
		b := NewBoard(w, h)
		for i, r := range scratch.glyphAt[:wh] {
			b.Data[i] = Cell{R: r}
		}
		if err := ValidatePartialBoard(b); err != nil {
			continue
		}
		heads = heads[:0]
		for _, p := range paths {
			heads = append(heads, p[0])
		}
		if !growPlayfulEnoughHeads(b, heads) {
			continue
		}
		if !verifySolvableFastBuf(b, heads, scratch) {
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
	var heads []Point
	for y := 0; y < b.H; y++ {
		for x := 0; x < b.W; x++ {
			if b.At(x, y).IsHead() {
				heads = append(heads, Point{x, y})
			}
		}
	}
	return growPlayfulEnoughHeads(b, heads)
}

// growPlayfulEnoughHeads is the same predicate as growPlayfulEnough but skips the
// full-board scan when the caller already knows the head positions.
func growPlayfulEnoughHeads(b Board, heads []Point) bool {
	total := len(heads)
	if total <= 1 {
		return true
	}
	fireable := 0
	for _, h := range heads {
		if RayEscapes(b, h.X, h.Y) {
			fireable++
		}
	}
	return 2*fireable <= total
}
