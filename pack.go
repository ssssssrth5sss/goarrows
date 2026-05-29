package main

import (
	"goarrows/game"
	"goarrows/levels"
)

// loadPack builds a procedural pack and ensures level 0 can be generated (fail fast).
func loadPack(seed int64) (*levels.Pack, error) {
	p := levels.NewProceduralPack(seed)
	if _, _, err := p.LevelAt(0); err != nil {
		return nil, err
	}
	return p, nil
}

// newGameForLevel loads level idx; negative startLives is mapped to a huge life pool for “unlimited”.
func newGameForLevel(p *levels.Pack, idx, startLives int) *game.Game {
	b, name, err := p.LevelAt(idx)
	if err != nil {
		panic(err)
	}
	lives := startLives
	if lives < 0 {
		lives = 1 << 30
	}
	return game.NewGame(b, lives, name)
}

// newGameWithGenOverlay shows a “generating” overlay if the procedural level is not cached yet.
func newGameWithGenOverlay(pack *levels.Pack, idx, startLives int, generatingN *int, redraw func()) *game.Game {
	n := pack.ProceduralSideLen(idx)
	if n > 0 && !pack.ProceduralLevelReady(idx) {
		*generatingN = n
		redraw()
	}
	g := newGameForLevel(pack, idx, startLives)
	*generatingN = 0
	return g
}

// resetLevel reloads the same index from the pack and resets lives (replay current level).
func resetLevel(p *levels.Pack, g **game.Game, idx, startLives int) {
	b, _, err := p.LevelAt(idx)
	if err != nil {
		panic(err)
	}
	lives := startLives
	if lives < 0 {
		lives = 1 << 30
	}
	(*g).Reset(b, lives)
}
