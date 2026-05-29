package main

import (
	"testing"

	"goarrows/levels"
)

// TestLoadPack_procedural ensures loadPack succeeds and yields a procedural pack of expected Len.
func TestLoadPack_procedural(t *testing.T) {
	p, err := loadPack(42)
	if err != nil {
		t.Fatalf("loadPack returned error: %v", err)
	}
	if p == nil {
		t.Fatal("loadPack returned nil pack")
	}
	if got, want := p.Len(), levels.ProceduralLevelCount; got != want {
		t.Fatalf("pack len=%d want %d", got, want)
	}
}
