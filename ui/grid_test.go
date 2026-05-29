package ui

import "testing"

// TestOverlayHasHorizontalEdge_UShapeLegsNotBridged guards the bug where a vertical
// U-shaped arrow's two legs (adjacent columns sharing rows) were joined by a spurious
// '─' bridge during fire animation, making the vertical segments look horizontal. Only
// genuinely consecutive polyline cells should bridge, so the legs stay unbridged while
// the real bottom edge does bridge.
func TestOverlayHasHorizontalEdge_UShapeLegsNotBridged(t *testing.T) {
	// North-firing U: head ▲(1,0) down the left leg, across the bottom, up the right leg.
	cells := []OverlayCell{
		{X: 1, Y: 0, R: '▲'},
		{X: 1, Y: 1, R: '│'},
		{X: 1, Y: 2, R: '└'},
		{X: 2, Y: 2, R: '┘'},
		{X: 2, Y: 1, R: '│'},
		{X: 2, Y: 0, R: '│'},
	}
	if overlayHasHorizontalEdge(cells, 1, 0) {
		t.Fatal("legs at row 0 must not be bridged")
	}
	if overlayHasHorizontalEdge(cells, 1, 1) {
		t.Fatal("legs at row 1 must not be bridged")
	}
	if !overlayHasHorizontalEdge(cells, 1, 2) {
		t.Fatal("the bottom └┘ edge must be bridged")
	}
}

// TestOverlayHasHorizontalEdge_OffBoardNoNeighborRowBridge guards the earlier aliasing bug:
// an arrow flying off the east edge must not report a bridge in the neighbor row.
func TestOverlayHasHorizontalEdge_OffBoardNoNeighborRowBridge(t *testing.T) {
	// East-firing arrow with head and leading body past the east edge.
	cells := []OverlayCell{
		{X: 4, Y: 0, R: '>'},
		{X: 3, Y: 0, R: '─'},
		{X: 2, Y: 0, R: '─'},
	}
	if overlayHasHorizontalEdge(cells, 0, 1) {
		t.Fatal("off-board cells must not produce a bridge in the neighbor row")
	}
	if !overlayHasHorizontalEdge(cells, 2, 0) {
		t.Fatal("the on-board horizontal segment must bridge")
	}
}
