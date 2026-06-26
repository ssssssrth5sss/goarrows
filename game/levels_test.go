package game

import (
	"testing"
)

// TestLevelsAt checks level sizing, naming, memoization, and Count for the generator.
func TestLevelsAt(t *testing.T) {
	lv := NewLevels(42)
	b, name, err := lv.At(0)
	if err != nil {
		t.Fatal(err)
	}
	if name == "" || b.W != 3 || b.H != 3 {
		t.Fatalf("level 0: name=%q board=%dx%d", name, b.W, b.H)
	}
	b2, _, err := lv.At(0)
	if err != nil {
		t.Fatal(err)
	}
	if b2.W != b.W || b2.H != b.H {
		t.Fatal("memo mismatch")
	}
	b3, name3, err := lv.At(2)
	if err != nil {
		t.Fatal(err)
	}
	if b3.W != 5 || b3.H != 5 {
		t.Fatalf("level 2 want 5×5 got %d×%d", b3.W, b3.H)
	}
	if name3 == "" {
		t.Fatal("empty name")
	}
	if lv.Count() != MaxLevels {
		t.Fatalf("Count: %d", lv.Count())
	}
}

// TestLevelsAt_negative rejects a negative level index.
func TestLevelsAt_negative(t *testing.T) {
	lv := NewLevels(1)
	if _, _, err := lv.At(-1); err == nil {
		t.Fatal("expected error for negative index")
	}
}
