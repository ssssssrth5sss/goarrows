package main

import (
	"testing"
	"time"

	"goarrows/game"
	"goarrows/ui"
)

// TestBuildPointerFrames_EastStraight checks frame cells for a horizontal straight-line clear.
func TestBuildPointerFrames_EastStraight(t *testing.T) {
	b := game.NewBoard(6, 3)
	b.Set(3, 1, game.Cell{R: '>'})
	b.Set(2, 1, game.Cell{R: '─'})
	b.Set(1, 1, game.Cell{R: '─'})

	path, err := game.PathFromHead(b, 3, 1)
	if err != nil {
		t.Fatal(err)
	}
	ray := fireTravelCells(b, 3, 1)
	frames, ok := buildPointerFrames(b, path, ray, '>')
	if !ok {
		t.Fatal("buildPointerFrames returned !ok")
	}
	if len(frames) != 4 {
		t.Fatalf("frame count got %d want 4", len(frames))
	}

	assertFrameCells(t, frames[0].Cells, []ui.OverlayCell{
		{X: 4, Y: 1, R: '>'},
		{X: 3, Y: 1, R: '─'},
		{X: 2, Y: 1, R: '─'},
	})
	assertFrameCells(t, frames[1].Cells, []ui.OverlayCell{
		{X: 5, Y: 1, R: '>'},
		{X: 4, Y: 1, R: '─'},
		{X: 3, Y: 1, R: '─'},
	})
	assertFrameCells(t, frames[2].Cells, []ui.OverlayCell{
		{X: 6, Y: 1, R: '>'},
		{X: 5, Y: 1, R: '─'},
		{X: 4, Y: 1, R: '─'},
	})
	assertFrameCells(t, frames[3].Cells, []ui.OverlayCell{
		{X: 7, Y: 1, R: '>'},
		{X: 6, Y: 1, R: '─'},
		{X: 5, Y: 1, R: '─'},
	})
}

// TestBuildPointerFrames_NorthStraight checks frame cells for a vertical straight-line clear.
func TestBuildPointerFrames_NorthStraight(t *testing.T) {
	b := game.NewBoard(4, 6)
	b.Set(2, 2, game.Cell{R: '^'})
	b.Set(2, 3, game.Cell{R: '│'})
	b.Set(2, 4, game.Cell{R: '│'})

	path, err := game.PathFromHead(b, 2, 2)
	if err != nil {
		t.Fatal(err)
	}
	ray := fireTravelCells(b, 2, 2)
	frames, ok := buildPointerFrames(b, path, ray, '^')
	if !ok {
		t.Fatal("buildPointerFrames returned !ok")
	}
	if len(frames) != 4 {
		t.Fatalf("frame count got %d want 4", len(frames))
	}
	assertFrameCells(t, frames[0].Cells, []ui.OverlayCell{
		{X: 2, Y: 1, R: '^'},
		{X: 2, Y: 2, R: '│'},
		{X: 2, Y: 3, R: '│'},
	})
	assertFrameCells(t, frames[1].Cells, []ui.OverlayCell{
		{X: 2, Y: 0, R: '^'},
		{X: 2, Y: 1, R: '│'},
		{X: 2, Y: 2, R: '│'},
	})
}

// TestBuildPointerFrames_BentPath checks early frames when the polyline turns before exiting.
func TestBuildPointerFrames_BentPath(t *testing.T) {
	b := game.NewBoard(6, 4)
	b.Set(3, 1, game.Cell{R: '>'})
	b.Set(2, 1, game.Cell{R: '┌'})
	b.Set(2, 2, game.Cell{R: '┘'})
	b.Set(1, 2, game.Cell{R: '─'})

	path, err := game.PathFromHead(b, 3, 1)
	if err != nil {
		t.Fatal(err)
	}
	ray := fireTravelCells(b, 3, 1)
	frames, ok := buildPointerFrames(b, path, ray, '>')
	if !ok {
		t.Fatal("buildPointerFrames returned !ok")
	}
	if len(frames) != 5 {
		t.Fatalf("frame count got %d want 5", len(frames))
	}
	assertFrameCells(t, frames[0].Cells, []ui.OverlayCell{
		{X: 4, Y: 1, R: '>'},
		{X: 3, Y: 1, R: '─'},
		{X: 2, Y: 1, R: '┌'},
		{X: 2, Y: 2, R: '┘'},
	})
	assertFrameCells(t, frames[1].Cells, []ui.OverlayCell{
		{X: 5, Y: 1, R: '>'},
		{X: 4, Y: 1, R: '─'},
		{X: 3, Y: 1, R: '─'},
		{X: 2, Y: 1, R: '┌'},
	})
}

// TestBuildPointerFrames_TailKeepsBendGlyph checks that as the snake's tail recedes onto a
// former corner cell, the animation keeps the cell's original board glyph (the corner) rather
// than straightening it; only the cell directly behind the head becomes a straight body rune.
func TestBuildPointerFrames_TailKeepsBendGlyph(t *testing.T) {
	b := game.NewBoard(7, 5)
	b.Set(4, 2, game.Cell{R: '>'})
	b.Set(3, 2, game.Cell{R: '─'})
	b.Set(2, 2, game.Cell{R: '┌'})
	b.Set(2, 3, game.Cell{R: '│'})
	b.Set(2, 4, game.Cell{R: '│'})

	path, err := game.PathFromHead(b, 4, 2)
	if err != nil {
		t.Fatal(err)
	}
	ray := fireTravelCells(b, 4, 2)
	frames, ok := buildPointerFrames(b, path, ray, '>')
	if !ok {
		t.Fatal("buildPointerFrames returned !ok")
	}
	if len(frames) < 2 {
		t.Fatalf("frame count got %d want >= 2", len(frames))
	}

	// Frame 0: the corner at (2,2) keeps its board glyph ┌.
	assertFrameCells(t, frames[0].Cells, []ui.OverlayCell{
		{X: 5, Y: 2, R: '>'},
		{X: 4, Y: 2, R: '─'},
		{X: 3, Y: 2, R: '─'},
		{X: 2, Y: 2, R: '┌'},
		{X: 2, Y: 3, R: '│'},
	})
	// Frame 1: the tail recedes onto (2,2); it retains its original corner glyph ┌ instead
	// of straightening, since the animation only rewrites the cell behind the head.
	assertFrameCells(t, frames[1].Cells, []ui.OverlayCell{
		{X: 6, Y: 2, R: '>'},
		{X: 5, Y: 2, R: '─'},
		{X: 4, Y: 2, R: '─'},
		{X: 3, Y: 2, R: '─'},
		{X: 2, Y: 2, R: '┌'},
	})
}

// TestHeadPositionForStep verifies headPositionForStep on-board and past-edge steps.
func TestHeadPositionForStep(t *testing.T) {
	ox, oy := 1, 1
	x, y := headPositionForStep(ox, oy, 1, 0, 1)
	if x != 2 || y != 1 {
		t.Fatalf("step1 got (%d,%d) want (2,1)", x, y)
	}
	x, y = headPositionForStep(ox, oy, 1, 0, 3)
	if x != 4 || y != 1 {
		t.Fatalf("step3 got (%d,%d) want (4,1)", x, y)
	}
}

// TestFireTravelCells_AllDirections checks open-ray cell lists for all four head directions.
func TestFireTravelCells_AllDirections(t *testing.T) {
	tests := []struct {
		name   string
		w, h   int
		hx, hy int
		rune   rune
		want   []game.Point
	}{
		{"east", 6, 3, 2, 1, '>', []game.Point{{X: 3, Y: 1}, {X: 4, Y: 1}, {X: 5, Y: 1}}},
		{"west", 6, 3, 3, 1, '<', []game.Point{{X: 2, Y: 1}, {X: 1, Y: 1}, {X: 0, Y: 1}}},
		{"north", 4, 6, 2, 3, '^', []game.Point{{X: 2, Y: 2}, {X: 2, Y: 1}, {X: 2, Y: 0}}},
		{"south", 4, 6, 1, 2, 'v', []game.Point{{X: 1, Y: 3}, {X: 1, Y: 4}, {X: 1, Y: 5}}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			b := game.NewBoard(tc.w, tc.h)
			b.Set(tc.hx, tc.hy, game.Cell{R: tc.rune})
			got := fireTravelCells(b, tc.hx, tc.hy)
			assertPoints(t, got, tc.want)
		})
	}
}

// TestTryStartFireAnimation_InitializesState checks anim state after a successful start on a clear shot.
func TestTryStartFireAnimation_InitializesState(t *testing.T) {
	b := game.NewBoard(6, 3)
	b.Set(3, 1, game.Cell{R: '>'})
	b.Set(2, 1, game.Cell{R: '─'})
	b.Set(1, 1, game.Cell{R: '─'})
	g := game.NewGame(b, 3, "t")
	var anim animState
	ok := tryStartFireAnimation(g, 3, 1, &anim, 10*time.Millisecond)
	if !ok {
		t.Fatal("expected animation to start")
	}
	if !anim.active {
		t.Fatal("anim.active = false")
	}
	if len(anim.frames) == 0 {
		t.Fatal("anim.frames empty")
	}
	if anim.fireX != 3 || anim.fireY != 1 {
		t.Fatalf("fire origin got (%d,%d) want (3,1)", anim.fireX, anim.fireY)
	}
}

// TestTryStartFireAnimation_HeadAtEdge guards the bug where a head already on the board
// edge fired off-board without animation: the escape ray is empty, yet the body must still
// slide off, so the animation must start instead of clearing the arrow instantly.
func TestTryStartFireAnimation_HeadAtEdge(t *testing.T) {
	b := game.NewBoard(4, 3)
	b.Set(3, 1, game.Cell{R: '>'})
	b.Set(2, 1, game.Cell{R: '─'})
	b.Set(1, 1, game.Cell{R: '─'})
	if cells := fireTravelCells(b, 3, 1); len(cells) != 0 {
		t.Fatalf("expected empty travel ray for edge head, got %d cells", len(cells))
	}
	g := game.NewGame(b, 3, "t")
	var anim animState
	ok := tryStartFireAnimation(g, 3, 1, &anim, 10*time.Millisecond)
	if !ok {
		t.Fatal("animation should start even when the head is on the board edge")
	}
	if !anim.active {
		t.Fatal("anim.active = false")
	}
	if len(anim.frames) == 0 {
		t.Fatal("anim.frames empty: edge head was cleared without animation")
	}
}

// TestBuildPointerFrames_EdgeNoRay checks frames when the head starts on the edge (empty ray):
// the head steps off-board immediately while the body trails one cell per step.
func TestBuildPointerFrames_EdgeNoRay(t *testing.T) {
	b := game.NewBoard(4, 3)
	b.Set(3, 1, game.Cell{R: '>'})
	b.Set(2, 1, game.Cell{R: '─'})
	b.Set(1, 1, game.Cell{R: '─'})

	path, err := game.PathFromHead(b, 3, 1)
	if err != nil {
		t.Fatal(err)
	}
	ray := fireTravelCells(b, 3, 1)
	if len(ray) != 0 {
		t.Fatalf("expected empty ray, got %d", len(ray))
	}
	frames, ok := buildPointerFrames(b, path, ray, '>')
	if !ok {
		t.Fatal("buildPointerFrames returned !ok for edge head")
	}
	if len(frames) != 2 {
		t.Fatalf("frame count got %d want 2", len(frames))
	}
	assertFrameCells(t, frames[0].Cells, []ui.OverlayCell{
		{X: 4, Y: 1, R: '>'},
		{X: 3, Y: 1, R: '─'},
		{X: 2, Y: 1, R: '─'},
	})
	assertFrameCells(t, frames[1].Cells, []ui.OverlayCell{
		{X: 5, Y: 1, R: '>'},
		{X: 4, Y: 1, R: '─'},
		{X: 3, Y: 1, R: '─'},
	})
}

// assertFrameCells fails the test if overlay cell slices differ element-wise.
func assertFrameCells(t *testing.T, got, want []ui.OverlayCell) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("cell len got %d want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("cell[%d] got (%d,%d,%q) want (%d,%d,%q)",
				i, got[i].X, got[i].Y, got[i].R, want[i].X, want[i].Y, want[i].R)
		}
	}
}

// assertPoints fails the test if coordinate slices differ element-wise.
func assertPoints(t *testing.T, got, want []game.Point) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("point len got %d want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("point[%d] got (%d,%d) want (%d,%d)", i, got[i].X, got[i].Y, want[i].X, want[i].Y)
		}
	}
}
