package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"goarrows/game"
	"goarrows/ui"
)

// main parses flags, opens the tcell screen, runs the input/render loop (HUD, help, modals,
// optional fire animation), and tears down on quit.
func main() {
	startLives := flag.Int("lives", 3, "starting lives per level (use -1 for unlimited)")
	seedFlag := &optionalInt64Flag{}
	flag.Var(seedFlag, "seed", "base RNG seed for procedural levels (omit for random from clock; -seed 0 fixes zero)")
	flag.Parse()

	seed := resolveProceduralSeed(seedFlag)
	lv, err := loadLevels(seed)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if *startLives < -1 || *startLives == 0 {
		fmt.Fprintln(os.Stderr, "lives must be >= 1 or -1 for unlimited")
		os.Exit(1)
	}

	s, err := tcell.NewScreen()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := s.Init(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer s.Fini()
	s.HideCursor()

	def := tcell.StyleDefault
	base := def.Foreground(tcell.ColorSilver).Background(tcell.ColorBlack)
	cursorSt := def.Foreground(tcell.ColorBlack).Background(tcell.ColorAqua).Bold(true)
	titleSt := def.Foreground(tcell.ColorYellow).Bold(true)
	msgSt := def.Foreground(tcell.ColorWhite)
	blockedSt := def.Foreground(tcell.ColorOrange).Bold(true)
	winSt := def.Foreground(tcell.ColorGreen).Bold(true)
	helpSt := def.Foreground(tcell.ColorGray)

	showHelp := false
	status := ""
	var modal *ui.ModalOverlay
	generatingN := 0
	var anim animState
	animStep := 75 * time.Millisecond

	idx := 0
	var g *game.Game
	cx, cy := 0, 0

	redraw := func() {
		s.Clear()
		sw, sh := s.Size()
		gw, gh := ui.GridSize(g.Board.W, g.Board.H)
		hudLines := 4
		oy := (sh - gh - hudLines) / 2
		if oy < 0 {
			oy = 0
		}
		ox := (sw - gw) / 2
		if ox < 0 {
			ox = 0
		}

		var fireAnim *ui.FireAnimOverlay
		if anim.active && anim.step < len(anim.frames) {
			fireAnim = &anim.frames[anim.step]
		}
		ui.DrawGrid(s, ox, oy, g.Board, cx, cy, base, cursorSt, fireAnim)

		lineY := oy + gh + 1
		if lineY >= sh {
			lineY = sh - 1
		}
		ui.DrawStr(s, 0, lineY, sw, fmt.Sprintf(" %s", g.LevelName), titleSt)
		lineY++
		if lineY < sh {
			livesStr := ui.FormatLives(g.Lives, *startLives)
			left := fmt.Sprintf(" Lives: %s   Cells: %d", livesStr, g.Board.NonEmptyCount())
			ui.DrawStr(s, 0, lineY, sw, left, msgSt)
		}
		lineY++
		if lineY < sh && status != "" {
			st := msgSt
			if strings.HasPrefix(status, "Blocked") {
				st = blockedSt
			} else if strings.HasPrefix(status, "Cleared") {
				st = winSt
			}
			ui.DrawStr(s, 0, lineY, sw, " "+status, st)
		}
		lineY++
		if lineY < sh {
			ui.DrawStr(s, 0, lineY, sw, " hjkl/←↑↓→ move  space/enter fire  r restart  n/p level  ? help  q quit", helpSt)
		}

		if showHelp {
			ui.DrawHelpOverlay(s, sw, sh, base)
		}
		if modal != nil {
			ui.DrawModalOverlay(s, sw, sh, modal, def)
		}
		if generatingN > 0 {
			ui.DrawGeneratingOverlay(s, sw, sh, generatingN, def)
		}
		s.Show()
	}
	clearTransient := func() {
		status, modal = "", nil
		anim.active = false
	}
	gotoLevel := func(delta int) {
		idx = (idx + delta + lv.Count()) % lv.Count()
		g = newGameWithGenOverlay(lv, idx, *startLives, &generatingN, redraw)
		clearTransient()
		clampCursor(g, &cx, &cy)
	}
	restart := func() {
		resetLevel(lv, &g, idx, *startLives)
		clearTransient()
		clampCursor(g, &cx, &cy)
	}
	tryStartOrApplyFire := func(x, y int) {
		started := tryStartFireAnimation(g, x, y, &anim, animStep)
		if !started {
			fr := applyFire(g, x, y, *startLives)
			status, modal = fr.status, fr.overlay
		}
	}
	interruptAndFire := func(x, y int) {
		if anim.active {
			fr := applyFire(g, anim.fireX, anim.fireY, *startLives)
			status, modal = fr.status, fr.overlay
			anim.active = false
		}
		if g.Won() || g.Lost() {
			return
		}
		tryStartOrApplyFire(x, y)
	}

	g = newGameWithGenOverlay(lv, idx, *startLives, &generatingN, redraw)
	clampCursor(g, &cx, &cy)
	redraw()
	ticker := time.NewTicker(16 * time.Millisecond)
	defer ticker.Stop()
	go func() {
		for range ticker.C {
			s.PostEvent(tcell.NewEventInterrupt(nil))
		}
	}()

	quit := false
	for !quit {
		ev := s.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventResize:
			s.Sync()
			redraw()
		case *tcell.EventInterrupt:
			if !anim.active || time.Now().Before(anim.nextStep) {
				continue
			}
			anim.step++
			anim.nextStep = time.Now().Add(animStep)
			if anim.step >= len(anim.frames) {
				anim.active = false
				fr := applyFire(g, anim.fireX, anim.fireY, *startLives)
				status, modal = fr.status, fr.overlay
			}
			redraw()
		case *tcell.EventKey:
			if showHelp {
				showHelp = false
				redraw()
				continue
			}
			if g.Won() || g.Lost() {
				switch ev.Key() {
				case tcell.KeyCtrlC, tcell.KeyEscape:
					quit = true
				case tcell.KeyEnter:
					if g.Won() {
						gotoLevel(1)
					}
				case tcell.KeyRune:
					switch ev.Rune() {
					case 'q', 'Q':
						quit = true
					case 'r', 'R':
						restart()
					case 'n', 'N':
						gotoLevel(1)
					case 'p', 'P':
						gotoLevel(-1)
					}
				}
				redraw()
				continue
			}

			switch ev.Key() {
			case tcell.KeyCtrlC:
				quit = true
			case tcell.KeyUp:
				moveCursor(g, &cx, &cy, 0, -1)
			case tcell.KeyDown:
				moveCursor(g, &cx, &cy, 0, 1)
			case tcell.KeyLeft:
				moveCursor(g, &cx, &cy, -1, 0)
			case tcell.KeyRight:
				moveCursor(g, &cx, &cy, 1, 0)
			case tcell.KeyEnter:
				interruptAndFire(cx, cy)
			case tcell.KeyRune:
				switch r := ev.Rune(); r {
				case 'q', 'Q':
					quit = true
				case 'h':
					moveCursor(g, &cx, &cy, -1, 0)
				case 'l':
					moveCursor(g, &cx, &cy, 1, 0)
				case 'k':
					moveCursor(g, &cx, &cy, 0, -1)
				case 'j':
					moveCursor(g, &cx, &cy, 0, 1)
				case ' ', 'f', 'F':
					interruptAndFire(cx, cy)
				case 'r', 'R':
					restart()
				case 'n', 'N':
					gotoLevel(1)
				case 'p', 'P':
					gotoLevel(-1)
				case '?':
					showHelp = true
				}
			}
			clampCursor(g, &cx, &cy)
			redraw()
		}
	}
}
