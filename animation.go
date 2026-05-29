package main

import (
	"time"

	"goarrows/game"
	"goarrows/ui"
)

type animState struct {
	active   bool
	hidePath []game.Point         // original fired path (masked during animation)
	frames   []ui.FireAnimOverlay // precomputed snake frames
	step     int
	nextStep time.Time
	fireX    int
	fireY    int
}

// tryStartFireAnimation prepares ray-snake frames for a clearing shot; returns false if animation is skipped.
func tryStartFireAnimation(g *game.Game, cx, cy int, anim *animState, stepDur time.Duration) bool {
	if g.Won() || g.Lost() || !g.Board.InBounds(cx, cy) {
		return false
	}
	c := g.Board.At(cx, cy)
	if c.IsEmpty() || !c.IsHead() || !game.RayEscapes(g.Board, cx, cy) {
		return false
	}
	path, err := game.PathFromHead(g.Board, cx, cy)
	if err != nil || len(path) == 0 {
		return false
	}
	// An edge head has no travel cells (cells is empty); the body must still
	// slide off, so we do not bail here. buildPointerFrames handles an empty ray.
	cells := fireTravelCells(g.Board, cx, cy)
	frames, ok := buildPointerFrames(g.Board, path, cells, c.R)
	if !ok || len(frames) == 0 {
		return false
	}
	anim.active = true
	anim.hidePath = path
	anim.frames = frames
	anim.step = 0
	anim.nextStep = time.Now().Add(stepDur)
	anim.fireX = cx
	anim.fireY = cy
	return true
}

// buildPointerFrames builds per-step overlays: head slides along ray then past the edge while the tail follows.
func buildPointerFrames(b game.Board, path, ray []game.Point, headRune rune) ([]ui.FireAnimOverlay, bool) {
	if len(path) == 0 {
		return nil, false
	}
	fireDir, ok := game.HeadFireDir(headRune)
	if !ok {
		return nil, false
	}
	dx, dy := game.Delta(fireDir)
	bodyRune := straightBodyRune(fireDir)
	ox, oy := path[0].X, path[0].Y
	cur := make([]ui.OverlayCell, len(path))
	for i, p := range path {
		cur[i] = ui.OverlayCell{X: p.X, Y: p.Y, R: b.At(p.X, p.Y).R}
	}
	// Keep animating after head reaches the boundary cell so the tail also
	// reaches and exits the boundary before we commit final clear.
	totalSteps := len(ray) + len(path) - 1
	frames := make([]ui.FireAnimOverlay, 0, totalSteps)
	for step := 1; step <= totalSteps; step++ {
		if len(cur) == 0 {
			break
		}
		cur[0].R = bodyRune
		hx, hy := headPositionForStep(ox, oy, dx, dy, step)
		next := ui.OverlayCell{X: hx, Y: hy, R: headRune}
		nxt := make([]ui.OverlayCell, 0, len(cur))
		nxt = append(nxt, next)
		if len(cur) > 1 {
			nxt = append(nxt, cur[:len(cur)-1]...)
		}
		cur = nxt
		frameCells := make([]ui.OverlayCell, len(cur))
		copy(frameCells, cur)
		frames = append(frames, ui.FireAnimOverlay{
			HidePath: path,
			Cells:    frameCells,
		})
	}
	return frames, len(frames) > 0
}

// headPositionForStep is the animated head cell after step steps (1-based) measured from
// the head origin (ox, oy) along the fire direction. The escape ray is straight, so this
// covers both on-board and off-board steps, including an edge head with no travel cells.
func headPositionForStep(ox, oy, dx, dy, step int) (int, int) {
	return ox + step*dx, oy + step*dy
}

// straightBodyRune is the wire rune left behind as the head moves (matches fire axis).
func straightBodyRune(d game.Direction) rune {
	switch d {
	case game.North, game.South:
		return '│'
	default:
		return '─'
	}
}

// fireTravelCells lists empty cells from the head cell along the open ray to the board edge (exclusive of head).
func fireTravelCells(b game.Board, cx, cy int) []game.Point {
	c := b.At(cx, cy)
	fire, ok := game.HeadFireDir(c.R)
	if !ok {
		return nil
	}
	dx, dy := game.Delta(fire)
	var out []game.Point
	for x, y := cx+dx, cy+dy; b.InBounds(x, y); x, y = x+dx, y+dy {
		out = append(out, game.Point{X: x, Y: y})
	}
	return out
}
