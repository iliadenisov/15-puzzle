package puzzle

import (
	"15-puzzle/internal/model"
	"errors"
	"fmt"
	"image"
	"image/color"
	"math/rand/v2"
	"sync/atomic"
	"time"
)

const (
	tiles     = 16
	fieldSymX = 29
	fieldSymY = 12
)

var (
	checkbox = map[bool]string{true: "[x]", false: "[ ]"}
	frames   = [...]byte{'|', '/', '-', '\\'}
)

type game struct {
	langCode     langCode
	tiles        [tiles]*tile
	moves        int
	solved       bool
	muted        bool
	onStart      func()
	onSolve      func(int)
	requestStats func()

	stats     atomic.Value
	blinkCoef []float64

	blink [fieldSymX][fieldSymY]int
	color [fieldSymX][fieldSymY]color.RGBA
}

func newGame(onStart func(), onSolve func(int), request func()) *game {
	p := &game{
		langCode:     langCodeEn,
		solved:       true,
		onStart:      onStart,
		onSolve:      onSolve,
		requestStats: request,
		blinkCoef:    []float64{1, .8, .6, .4, .2, 0, 0, .2, .4, .6, .8, 1},
	}

	for i := range tiles {
		var pos int
		if i == 0 {
			pos = tiles - 1
		} else {
			pos = i - 1
		}
		p.tiles[i] = NewTile(i, pos)
	}

	for x := range fieldSymX {
		for y := range fieldSymY {
			p.blink[x][y] = rand.IntN(len(frames))
		}
	}

	p.requestStats()

	return p
}

func (g *game) Tick(t ticker) {
	if t == ticker10Hz && g.solved {
		b := g.blinkCoef[0]
		for i := 1; i < len(g.blinkCoef); i++ {
			g.blinkCoef[i-1] = g.blinkCoef[i]
		}
		g.blinkCoef[len(g.blinkCoef)-1] = b
		if g.showCongrats() {
			for x := range fieldSymX {
				for y := range fieldSymY {
					g.blink[x][y] = (g.blink[x][y] + 1) % len(frames)
					g.color[x][y] = color.RGBA{uint8(rand.IntN(256)), uint8(rand.IntN(256)), uint8(rand.IntN(256)), 0xFF}
				}
			}
		}
	}
}

func (g *game) SetLang(lc langCode) {
	g.langCode = lc
}

func (g *game) ApiStatsHandler(s model.Stats) {
	g.stats.Store(s)
}

func (g *game) Draw(s Screen) {
	drawGameField(s)
	if g.isSolved() {
		rating := l10nRank(g.langCode) + ": "
		wins := l10nWins(g.langCode) + ": "
		if !g.withStats(func(s model.Stats) {
			if s.Rank > 0 && s.GamesSolved > 0 {
				rating += fmt.Sprint(s.Rank)
			} else {
				rating = ""
			}
			wins += fmt.Sprint(s.GamesSolved)
		}) {
			rating, wins = "_", "_"
		}
		printHeader(s, rating, -1)
		printHeader(s, wins, 1)
	} else {
		printHeader(s, fmt.Sprintf(l10nSilent(g.langCode)+" %s", checkbox[g.muted]), 1)
		if !g.muted {
			printHeader(s, fmt.Sprintf(l10nMoves(g.langCode)+": %d", g.moves), -1)
		}
	}

	if g.showCongrats() {
		for x := range fieldSymX {
			for y := range fieldSymY {
				s.Print(string(frames[g.blink[x][y]]), image.Point{x + 1, y + 3}, g.color[x][y])
			}
		}
		return
	}

	boardBottomColor := color.Black
	// boardBottomColor := color.RGBA{0xFF, 0x00, 0xFF, 0xFF}

	for i := range tiles {
		var tileBackground color.Color = color.RGBA{0, 0, 0xA0, 0xFF}    // background: 0x0000A0 (lighter) / 0x00006B (darker)
		var tileForeground color.Color = color.RGBA{0, 0xFF, 0xFF, 0xFF} // text: cyan 0x00FFFF
		fillRect := image.Rect(g.tiles[i].X(), g.tiles[i].Y(), g.tiles[i].X()+puzzleTileSymW, g.tiles[i].Y()+puzzleTileSymH)
		if i == 0 {
			if g.solved {
				cr, cg, cb, ca := tileForeground.RGBA()
				tileForeground = color.RGBA{byte(g.blinkCoef[0] * float64(cr)),
					byte(g.blinkCoef[0] * float64(cg)),
					byte(g.blinkCoef[0] * float64(cb)),
					byte(g.blinkCoef[0] * float64(ca))}
			} else {
				tileBackground = boardBottomColor
				fillRect.Max.X--
			}
		}
		if g.tiles[i].moving {
			// fillRect contains unshifted yet coordinates, time to fix font gallucinations when animating adjacent tiles
			fillRect.Max.X--                   // right border: skip last column to justify on move
			s.Fill(fillRect, boardBottomColor) // partial "board bottom" from both sides
			fillRect.Max.X++                   // right border: restore
			fillRect.Min.X--                   // left border: extend to 1 col left to justify on move
		}
		s.Fill(fillRect, tileBackground)
		if i == 0 && g.solved || i > 0 {
			s.Print(tileTemplate.Format(g.tiles[i].Title()), image.Point{g.tiles[i].X(), g.tiles[i].Y()}, tileForeground)
		}
	}
}

func (g *game) Interact(a Audio, col, row int, t time.Duration) (result actionResult) {
	result = resultNone
	if !g.isSolved() && row == 1 && col < puzzleSymX-2 && col > puzzleSymX-13 {
		g.muted = !g.muted
		return
	}
	if g.showCongrats() {
		g.moves = 0
		return
	}
	for i := range tiles {
		if g.tiles[i].CanInteract(col, row) {
			if g.tiles[i].num == 0 && t > time.Second*3 {
				return resultSwitchForm
			}
			if g.press(a, g.tiles[i]) {
				break
			}
		}
	}
	return
}

func (g *game) shuffle() {
	rand.Shuffle(tiles, func(i, j int) { g.tiles[i].pos, g.tiles[j].pos = g.tiles[j].pos, g.tiles[i].pos })
	var puzzle [tiles]byte
	for i := range g.tiles {
		puzzle[i] = byte(g.tiles[i].pos)
	}
	if solvable, err := IsSolvable(puzzle); err == nil && !solvable {
		g.tiles[1].pos, g.tiles[2].pos = g.tiles[2].pos, g.tiles[1].pos
	}
	g.moves = 0
	g.solved = g.isSolved()
}

func (g *game) press(a Audio, t *tile) bool {
	if g.solved {
		if g.tiles[0].num == t.num {
			g.shuffle()
		}
		return true
	}
	exchange := func() { t.pos, g.tiles[0].pos = g.tiles[0].pos, t.pos; g.onMove(a) }
	switch {
	case t.Col() != g.tiles[0].Col() && t.Row() != g.tiles[0].Row():
		return false
	case t.Col() == g.tiles[0].Col() && g.tiles[0].Row()-t.Row() == 1:
		t.MoveDown(exchange)
	case t.Col() == g.tiles[0].Col() && t.Row()-g.tiles[0].Row() == 1:
		t.MoveUp(exchange)
	case t.Row() == g.tiles[0].Row() && g.tiles[0].Col()-t.Col() == 1:
		t.MoveRight(exchange)
	case t.Row() == g.tiles[0].Row() && t.Col()-g.tiles[0].Col() == 1:
		t.MoveLeft(exchange)
	}
	return true
}

func (g *game) onMove(a Audio) {
	if !g.muted {
		a.PlaySound()
	}
	if g.moves == 0 {
		g.onStart()
	}
	g.moves++
	solved := g.isSolved()
	if solved && !g.solved {
		g.onSolve(g.moves)
	}
	g.solved = solved
}

func (g *game) isSolved() bool {
	for i := 1; i < tiles; i++ {
		if g.tiles[i].pos != i-1 {
			return false
		}
	}
	return g.moves > 0
}

func (g *game) isTopRated() bool {
	firstPlace := false
	firstPlace = g.withStats(func(s model.Stats) { firstPlace = s.Rank == 1 })
	return firstPlace
}

func (g *game) showCongrats() bool {
	return g.isSolved() && g.isTopRated()
}

func (g *game) withStats(consumer func(model.Stats)) bool {
	if stats := g.stats.Load(); stats != nil {
		consumer(stats.(model.Stats))
		return true
	}
	return false
}

func (g *game) Activate() {}

func IsSolvable(puzzle [tiles]byte) (bool, error) {
	var x *int
	for i := range puzzle {
		if puzzle[i] == 0 {
			if x != nil {
				return false, errors.New("only one blank position allowed")
			}
			x = &i
		}
	}
	if x == nil {
		return false, errors.New("blank position required to be set")
	}
	blank := *x
	invCount := 0
	for i := range tiles - 1 {
		for j := i + 1; j < tiles; j++ {
			if puzzle[j] > 0 && puzzle[i] > 0 && puzzle[i] > puzzle[j] {
				invCount += 1
			}
		}
	}
	if (blank/4+1)%2 == 0 {
		return invCount%2 == 0, nil
	} else {
		return invCount%2 == 1, nil
	}
}
