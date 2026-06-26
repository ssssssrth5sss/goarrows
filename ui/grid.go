package ui

import (
	"github.com/gdamore/tcell/v2"
	"goarrows/game"
)

// OverlayCell is a transient glyph drawn over the grid at logical coordinates.
type OverlayCell struct {
	X, Y int
	R    rune
}

// FireAnimOverlay draws a frame of fire animation.
type FireAnimOverlay struct {
	HidePath []game.Point
	Cells    []OverlayCell
}

// GridSize returns terminal width and height for a w×h logical board: each
// logical column uses two screen columns (glyph at 2x, optional ─ bridge at 2x+1),
// so width is 2*w-1 and height is h.
func GridSize(w, h int) (gw, gh int) {
	if w <= 0 {
		return 0, h
	}
	return 2*w - 1, h
}

// DrawGrid paints the board at (ox, oy). Logical cell (x,y) is drawn at screen
// (ox+2*x, oy+y). Between (x,y) and (x+1,y), a '─' is drawn when the path links
// horizontally so lines stay visually continuous.
func DrawGrid(s tcell.Screen, ox, oy int, b game.Board, cursorX, cursorY int, base, cursor tcell.Style, fireAnim *FireAnimOverlay) {
	hidePath := fireAnimHidePath(fireAnim)
	hideSet := pathCellSet(b.W, hidePath)
	animCells := fireAnimCells(fireAnim)
	animPath := overlayCellPoints(animCells)

	for y := 0; y < b.H; y++ {
		for x := 0; x < b.W; x++ {
			st := base
			if x == cursorX && y == cursorY {
				st = cursor
			}
			r := DisplayRune(b.At(x, y))
			if _, masked := hideSet[y*b.W+x]; masked {
				r = ' '
			}
			s.SetContent(ox+2*x, oy+y, r, nil, st)
		}
		for x := 0; x+1 < b.W; x++ {
			st := base
			if y == cursorY && (x == cursorX || x+1 == cursorX) {
				st = cursor
			}
			ch := ' '
			if game.HorizontalLink(b, x, y) {
				ch = '─'
			}
			if hasEastEdge(hidePath, x, y) {
				ch = ' '
			}
			if hasEastEdge(animPath, x, y) {
				ch = '─'
			}
			s.SetContent(ox+2*x+1, oy+y, ch, nil, st)
		}
	}
	for _, c := range animCells {
		if !b.InBounds(c.X, c.Y) {
			continue
		}
		st := base
		if c.X == cursorX && c.Y == cursorY {
			st = cursor
		}
		s.SetContent(ox+2*c.X, oy+c.Y, c.R, nil, st)
	}
}

// fireAnimHidePath returns cells masked to spaces during the fire animation (nil-safe).
func fireAnimHidePath(f *FireAnimOverlay) []game.Point {
	if f == nil {
		return nil
	}
	return f.HidePath
}

// fireAnimCells returns overlay glyphs for the current animation frame (nil-safe).
func fireAnimCells(f *FireAnimOverlay) []OverlayCell {
	if f == nil {
		return nil
	}
	return f.Cells
}

// pathCellSet maps linear cell indices (y*w+x) for fast membership on hidePath.
func pathCellSet(w int, path []game.Point) map[int]struct{} {
	m := make(map[int]struct{}, len(path))
	for _, p := range path {
		m[p.Y*w+p.X] = struct{}{}
	}
	return m
}

// overlayCellPoints extracts the logical coordinates of overlay cells in order.
func overlayCellPoints(cells []OverlayCell) []game.Point {
	pts := make([]game.Point, len(cells))
	for i, c := range cells {
		pts[i] = game.Point{X: c.X, Y: c.Y}
	}
	return pts
}

// hasEastEdge reports whether two consecutive points form the east edge between (x,y) and
// (x+1,y). Only adjacent polyline points bridge, so two separate legs that merely share a row
// (e.g. the sides of a U) are never joined.
func hasEastEdge(pts []game.Point, x, y int) bool {
	for i := 0; i+1 < len(pts); i++ {
		a, b := pts[i], pts[i+1]
		if (a.X == x && a.Y == y && b.X == x+1 && b.Y == y) ||
			(b.X == x && b.Y == y && a.X == x+1 && a.Y == y) {
			return true
		}
	}
	return false
}

// DisplayRune returns the glyph to draw for a cell (space if empty).
func DisplayRune(c game.Cell) rune {
	if c.IsEmpty() {
		return ' '
	}
	return c.R
}
