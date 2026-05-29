package ui

import "testing"

// TestOverlayCellSet_OffBoardCellsNoNeighborRowBridge guards the bug where an arrow
// flying almost off the board drew a spurious dashed line in the neighbor row: an
// off-board cell's y*w+x index aliased onto an in-bounds cell one row down, so the
// bridge logic painted '─' there.
func TestOverlayCellSet_OffBoardCellsNoNeighborRowBridge(t *testing.T) {
	w, h := 4, 3
	// East-firing arrow nearly off the board: head + leading body past the east edge.
	cells := []OverlayCell{
		{X: 4, Y: 0, R: '>'}, // off-board; 0*4+4 aliases onto (0,1)
		{X: 5, Y: 0, R: '─'}, // off-board; aliases onto (1,1)
		{X: 3, Y: 0, R: '─'}, // last on-board cell
	}
	set := overlayCellSet(w, h, cells)
	if hasHorizontalOverlayEdge(set, w, 0, 1) {
		t.Fatal("off-board cells aliased into row 1 and drew a spurious bridge")
	}
}
