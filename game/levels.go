package game

import (
	"fmt"
	"math/rand/v2"
)

// MaxLevels is the nominal number of levels, used for UI wraparound and titles.
const MaxLevels = 1 << 20

// levelGenMaxTries is how many distinct base seeds (seed, seed+1, …) we try per level
// when GenerateBoard fails, before giving up.
const levelGenMaxTries = 512

type levelMemo struct {
	b    Board
	name string
	err  error
}

// Levels generates puzzle boards on demand: level i is an (i+3)×(i+3) board,
// deterministic per base seed and cached after its first build.
type Levels struct {
	seed int64
	memo map[int]levelMemo
}

// NewLevels creates an on-demand, memoizing level generator for the given base seed.
func NewLevels(seed int64) *Levels {
	return &Levels{
		seed: seed,
		memo: make(map[int]levelMemo),
	}
}

// Count returns the nominal number of levels (used for UI modulo and titles).
func (l *Levels) Count() int {
	return MaxLevels
}

// SideLen returns the N in N×N for level index i.
func (l *Levels) SideLen(i int) int {
	return i + 3
}

// Ready reports whether level i is already generated and cached.
func (l *Levels) Ready(i int) bool {
	_, ok := l.memo[i]
	return ok
}

// At builds or returns cached level i: an (i+3)×(i+3) board, trying successive RNG seeds on failure.
func (l *Levels) At(i int) (Board, string, error) {
	if i < 0 {
		return Board{}, "", fmt.Errorf("negative level index")
	}
	if m, hit := l.memo[i]; hit {
		if m.err != nil {
			return Board{}, m.name, m.err
		}
		return m.b, m.name, nil
	}
	n := i + 3
	name := fmt.Sprintf("Level %d (%d×%d)", i+1, n, n)
	var b Board
	var err error
	for delta := int64(0); delta < levelGenMaxTries; delta++ {
		rng := levelRNG(l.seed+delta, i)
		b, err = GenerateBoard(n, n, rng)
		if err == nil {
			l.memo[i] = levelMemo{b: b, name: name}
			return b, name, nil
		}
	}
	l.memo[i] = levelMemo{name: name, err: err}
	return Board{}, name, err
}

// levelRNG is deterministic for a given (seed, level index).
func levelRNG(seed int64, idx int) *rand.Rand {
	s0 := uint64(seed) ^ uint64(uint32(idx))*0x9E3779B1
	s1 := uint64(idx)*0xC6A4A7935BD1E995 + uint64(seed)
	if s1%2 == 0 {
		s1++
	}
	return rand.New(rand.NewPCG(s0, s1))
}
