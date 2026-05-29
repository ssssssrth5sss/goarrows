package game

import (
	"fmt"
)

// ValidateBoard checks full coverage, mutual links, path shape, and one head per component.
func ValidateBoard(b Board) error {
	for y := 0; y < b.H; y++ {
		for x := 0; x < b.W; x++ {
			c := b.At(x, y)
			if c.IsEmpty() {
				return fmt.Errorf("cell (%d,%d): empty cell not allowed (full coverage)", x, y)
			}
			if !c.IsHead() && NominalPorts(c.R) == 0 {
				return fmt.Errorf("cell (%d,%d): unknown rune %q", x, y, c.R)
			}
		}
	}
	return checkPortsAndComponents(b, false)
}

// ValidatePartialBoard checks mutual links, path shape, and one head per connected component
// for boards that may contain empty cells (background). Empty cells are ignored.
// Full-coverage levels should use ValidateBoard instead.
func ValidatePartialBoard(b Board) error {
	hasAny := false
	for y := 0; y < b.H; y++ {
		for x := 0; x < b.W; x++ {
			c := b.At(x, y)
			if c.IsEmpty() {
				continue
			}
			hasAny = true
			if !c.IsHead() && NominalPorts(c.R) == 0 {
				return fmt.Errorf("cell (%d,%d): unknown rune %q", x, y, c.R)
			}
		}
	}
	if !hasAny {
		return fmt.Errorf("partial board: no non-empty cells")
	}
	return checkPortsAndComponents(b, true)
}

// checkPortsAndComponents verifies port degrees (heads link once; wires have degree 1 or 2) and
// that every connected component holds exactly one head. When skipEmpty is true, empty background
// cells are ignored; otherwise the caller has already guaranteed full coverage.
func checkPortsAndComponents(b Board, skipEmpty bool) error {
	for y := 0; y < b.H; y++ {
		for x := 0; x < b.W; x++ {
			c := b.At(x, y)
			if skipEmpty && c.IsEmpty() {
				continue
			}
			bits := popcount(EffectivePorts(b, x, y))
			if c.IsHead() {
				if bits != 1 {
					return fmt.Errorf("cell (%d,%d): head must have exactly one path link (got %d)", x, y, bits)
				}
				continue
			}
			if bits != 1 && bits != 2 {
				return fmt.Errorf("cell (%d,%d): wire must have degree 1 (tail) or 2 (internal), got %d", x, y, bits)
			}
		}
	}

	seen := make([]bool, b.W*b.H)
	for y := 0; y < b.H; y++ {
		for x := 0; x < b.W; x++ {
			idx := y*b.W + x
			if seen[idx] || (skipEmpty && b.At(x, y).IsEmpty()) {
				continue
			}
			heads := countComponentHeads(b, x, y, seen)
			if heads != 1 {
				return fmt.Errorf("component starting at (%d,%d): want exactly 1 head, got %d", x, y, heads)
			}
		}
	}
	return nil
}

// countComponentHeads flood-fills the component containing (sx,sy) along effective ports,
// marking visited cells in seen, and returns how many heads it contains.
func countComponentHeads(b Board, sx, sy int, seen []bool) int {
	heads := 0
	queue := []Point{{sx, sy}}
	seen[sy*b.W+sx] = true
	for qi := 0; qi < len(queue); qi++ {
		cx, cy := queue[qi].X, queue[qi].Y
		if b.At(cx, cy).IsHead() {
			heads++
		}
		eff := EffectivePorts(b, cx, cy)
		try := func(nx, ny int, port uint8) {
			if eff&port == 0 {
				return
			}
			nidx := ny*b.W + nx
			if seen[nidx] {
				return
			}
			seen[nidx] = true
			queue = append(queue, Point{nx, ny})
		}
		try(cx, cy-1, PortN)
		try(cx+1, cy, PortE)
		try(cx, cy+1, PortS)
		try(cx-1, cy, PortW)
	}
	return heads
}
