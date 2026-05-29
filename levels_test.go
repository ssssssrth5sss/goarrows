package main

import (
	"testing"

	"goarrows/game"
)

// TestLoadLevels_procedural ensures loadLevels succeeds and yields a generator of expected Count.
func TestLoadLevels_procedural(t *testing.T) {
	lv, err := loadLevels(42)
	if err != nil {
		t.Fatalf("loadLevels returned error: %v", err)
	}
	if lv == nil {
		t.Fatal("loadLevels returned nil generator")
	}
	if got, want := lv.Count(), game.MaxLevels; got != want {
		t.Fatalf("Count=%d want %d", got, want)
	}
}
