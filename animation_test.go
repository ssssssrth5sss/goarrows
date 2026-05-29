package main

import (
	"testing"
	"time"

	"goarrows/game"
)

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
