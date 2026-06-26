package main

import (
	"goarrows/game"
)

// clampCursor keeps the cursor inside the board rectangle.
func clampCursor(g *game.Game, cx, cy *int) {
	if *cx >= g.Board.W {
		*cx = g.Board.W - 1
	}
	if *cy >= g.Board.H {
		*cy = g.Board.H - 1
	}
	if *cx < 0 {
		*cx = 0
	}
	if *cy < 0 {
		*cy = 0
	}
}

// moveCursor applies a delta in logical cells and clamps to the board.
func moveCursor(g *game.Game, cx, cy *int, dx, dy int) {
	*cx += dx
	*cy += dy
	clampCursor(g, cx, cy)
}
