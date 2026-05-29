package game

import (
	"fmt"
)

// PathFromHead returns all cells on the polyline containing the head at (hx,hy), including the head.
func PathFromHead(b Board, hx, hy int) ([]Point, error) {
	if !b.InBounds(hx, hy) || !b.At(hx, hy).IsHead() {
		return nil, fmt.Errorf("not a head at (%d,%d)", hx, hy)
	}
	heff := EffectivePorts(b, hx, hy)
	if popcount(heff) != 1 {
		return nil, fmt.Errorf("head at (%d,%d): expected one body link", hx, hy)
	}
	bx, by, ok := stepAlongPorts(b, hx, hy, heff, -1, -1)
	if !ok {
		return nil, fmt.Errorf("head at (%d,%d): no body neighbor", hx, hy)
	}

	out := []Point{{hx, hy}}
	px, py := hx, hy
	cx, cy := bx, by
	for {
		out = append(out, Point{cx, cy})
		eff := EffectivePorts(b, cx, cy)
		var nx, ny int
		var found bool
		try := func(tx, ty int, needPort uint8) {
			if eff&needPort == 0 || found {
				return
			}
			if tx == px && ty == py {
				return
			}
			if !b.InBounds(tx, ty) {
				return
			}
			dir := directionFromTo(cx, cy, tx, ty)
			if !linked(b.At(cx, cy), b.At(tx, ty), dir) {
				return
			}
			nx, ny = tx, ty
			found = true
		}
		try(cx, cy-1, PortN)
		try(cx+1, cy, PortE)
		try(cx, cy+1, PortS)
		try(cx-1, cy, PortW)
		if !found {
			break
		}
		px, py, cx, cy = cx, cy, nx, ny
	}
	return out, nil
}

// stepAlongPorts picks a neighbor reachable from (x,y) along eff, excluding (px,py).
func stepAlongPorts(b Board, x, y int, eff uint8, px, py int) (int, int, bool) {
	cands := []struct {
		tx, ty int
		port   uint8
	}{
		{x, y - 1, PortN},
		{x + 1, y, PortE},
		{x, y + 1, PortS},
		{x - 1, y, PortW},
	}
	for _, c := range cands {
		if eff&c.port == 0 {
			continue
		}
		if c.tx == px && c.ty == py {
			continue
		}
		if !b.InBounds(c.tx, c.ty) {
			continue
		}
		dir := directionFromTo(x, y, c.tx, c.ty)
		if !linked(b.At(x, y), b.At(c.tx, c.ty), dir) {
			continue
		}
		return c.tx, c.ty, true
	}
	return 0, 0, false
}
