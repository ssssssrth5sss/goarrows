package main

import (
	"goarrows/game"
)

// loadLevels builds the level generator and ensures level 0 can be generated (fail fast).
func loadLevels(seed int64) (*game.Levels, error) {
	lv := game.NewLevels(seed)
	if _, _, err := lv.At(0); err != nil {
		return nil, err
	}
	return lv, nil
}

// newGameForLevel loads level idx; negative startLives is mapped to a huge life pool for “unlimited”.
func newGameForLevel(lv *game.Levels, idx, startLives int) *game.Game {
	b, name, err := lv.At(idx)
	if err != nil {
		panic(err)
	}
	lives := startLives
	if lives < 0 {
		lives = 1 << 30
	}
	return game.NewGame(b, lives, name)
}

// newGameWithGenOverlay shows a “generating” overlay if the level is not cached yet.
func newGameWithGenOverlay(lv *game.Levels, idx, startLives int, generatingN *int, redraw func()) *game.Game {
	if !lv.Ready(idx) {
		*generatingN = lv.SideLen(idx)
		redraw()
	}
	g := newGameForLevel(lv, idx, startLives)
	*generatingN = 0
	return g
}

// resetLevel reloads the same index and resets lives (replay current level).
func resetLevel(lv *game.Levels, g **game.Game, idx, startLives int) {
	b, _, err := lv.At(idx)
	if err != nil {
		panic(err)
	}
	lives := startLives
	if lives < 0 {
		lives = 1 << 30
	}
	(*g).Reset(b, lives)
}
