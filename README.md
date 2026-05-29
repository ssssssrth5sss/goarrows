# Go Arrows

A tiny puzzle game you play in your terminal.

## What is this?

The screen fills up with arrows. Each arrow points in a direction. When you
fire an arrow, it flies in the direction it points:

- If its path to the edge of the screen is clear, the arrow flies off and
  disappears.
- If another arrow is in the way, the shot is blocked and you lose a life.

Clear every arrow off the board to win the level.

## How to play

You need [Go](https://go.dev/dl/) installed and a terminal that can show
symbols (most modern terminals can). Then run:

```bash
go run .
```

Move the cursor to an arrow, fire it, and watch it go. Keep clearing arrows
until the board is empty.

## Controls

You play entirely with the keyboard:

| Key | What it does |
|--------|--------|
| `h` `j` `k` `l` or arrow keys | Move the cursor |
| Space, Enter, or `f` | Fire the arrow under the cursor |
| `r` | Restart the current level |
| `n` / `p` | Go to the next / previous level |
| `?` | Show or hide the help screen |
| `q` or Ctrl+C | Quit |

After you win or run out of lives, the bottom of the screen tells you which
keys to press next.

## Tips

- Look at where each arrow points before you fire.
- If an arrow is blocked, clear the arrows in its way first.
- Press `?` any time to see the help screen.
- Made a mistake? Press `r` to restart the level.

## Options

You can change a couple of things when you start the game:

- Lives: how many mistakes you can make per level (default is 3).

```bash
go run . -lives 5
```

  Use `-lives -1` for unlimited lives.

- Seed: the puzzles are made by the computer. Normally you get fresh, random
  puzzles every time. If you pass a seed number, you get the same set of
  puzzles every time, which is handy for replaying or sharing a challenge.

```bash
go run . -seed 42
```

## For developers

The game has no fixed level files; it builds puzzles as you play.

How levels are generated:

- Level *k* (shown in the on-screen info) is a `(k+2)x(k+2)` grid, so level 1
  is 3x3, level 2 is 4x4, and so on.
- The "grow" generator places a few small arrows and randomly extends them
  until they get stuck. Some cells may be left empty.
- A generated board is only accepted if it is not too easy (at most half the
  arrows have a clear shot to start) and if it can actually be solved by
  repeatedly firing the first available arrow. This guarantees every level is
  winnable.
- With `-seed`, generation is deterministic, so the same seed always produces
  the same levels. Levels are built when you reach them and cached for the run.

Code layout:

| Package | Role |
|---------|------|
| `main` | [tcell](https://github.com/gdamore/tcell) setup, the input loop, the on-screen info (level name, lives, remaining cells), status messages, and the help overlay. |
| `game` | The core rules: the board model, parsing, validation, `PathFromHead`, `TryFire` / `RayEscapes`, and level generation (`GenerateBoard`, `ValidatePartialBoard`, `VerifyGreedyFirstClearsBoard`, plus `VerifySolvable` used in tests). |
| `levels` | `NewProceduralPack(seed)` and `Pack.LevelAt` provide generated boards on demand. |
| `ui` | `DrawGrid` draws the board, placing each cell at screen column `2*x` and joining horizontal wires with `─` so they read as one line; `GridSize` is `(2*w-1, h)`. |

Where things live (one responsibility per file):

- `main` (root): `main.go` (event/render loop), `flags.go` (CLI flags and
  seed resolution), `pack.go` (pack/game glue), `cursor.go` (cursor movement),
  `fire.go` (fire outcome to status/modal), `animation.go` (fire-animation
  frames).
- `game`: `board.go` (the board model), `game.go` (game state plus `TryFire` /
  `RayEscapes`), `ports.go` (the connection model), `validate.go`, `path.go`,
  `level.go` (parsing), `gen.go` / `gen_grow.go` / `paint.go` (level
  generation), `solvable.go`.
- `levels`: `pack.go` (the `Pack` type), `procedural.go` (on-demand generation
  and memo).
- `ui`: `grid.go` (board drawing), `overlay.go` (HUD text and modals).

The game rules stay independent of the terminal: `game.TryFire` updates the
board and lives, while `main` only handles drawing and input.

## Status

Playable terminal game with automatically generated levels.
