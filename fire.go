package main

import (
	"goarrows/game"
	"goarrows/ui"
)

// fireUIResult combines an optional status line (e.g. Blocked) with an optional modal.
type fireUIResult struct {
	status  string
	overlay *ui.ModalOverlay
}

// applyFire runs TryFire and maps outcomes to status text and win/lose modals (no-op if already terminal).
func applyFire(g *game.Game, cx, cy, startLives int) fireUIResult {
	if g.Won() || g.Lost() {
		return fireUIResult{}
	}
	switch game.TryFire(g, cx, cy) {
	case game.FireNone:
		return fireUIResult{}
	case game.FireCleared:
		if g.Won() {
			return fireUIResult{
				overlay: &ui.ModalOverlay{
					Positive: true,
					Lines: []string{
						"You win!",
						"",
						"Press Enter for next level",
						"",
						"n next  p prev  r replay  q quit",
					},
				},
			}
		}
		return fireUIResult{status: "Cleared."}
	case game.FireBlocked:
		if g.Lost() {
			return fireUIResult{
				overlay: &ui.ModalOverlay{
					Positive: false,
					Lines: []string{
						"Game over",
						"",
						"r restart  q quit",
					},
				},
			}
		}
		return fireUIResult{status: "Blocked!"}
	default:
		return fireUIResult{}
	}
}
