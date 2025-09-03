package puzzle

import (
	"15-puzzle/internal/model"
	"fmt"
	"image"
	"image/color"
	"strconv"
	"sync/atomic"
	"time"
)

const (
	dials             = 10
	codeInputTemplate = `[%s] [%s] [%s] [%s]`
)

type form struct {
	dials       [dials]*button
	request     func(string)
	requestSent bool
	authFail    bool
	idx         int
	input       [4]byte
	mon         atomic.Value
}

func intFn(i int) func() int          { return func() int { return i } }
func stringFn(s string) func() string { return func() string { return s } }
func codeBox(b byte) string {
	if b > 0 {
		return "*"
	}
	return "_"
}

func newForm(request func(string)) *form {
	col := func(i int) int { return (i%3)*puzzleTileSymW + 5 }
	row := func(i int) int { return (i/3)*puzzleTileSymH + 3 }
	f := &form{request: request}
	for i := 1; i < dials; i++ {
		f.dials[i] = NewButton(stringFn(strconv.Itoa(i)), intFn(col(i-1)), intFn(row(i-1)))
	}
	f.dials[0] = NewButton(stringFn("0"), intFn(col(10)), intFn(row(10)))
	return f
}

func (f *form) ApiMonitoringHandler(m model.Monitoring) {
	f.mon.Store(m)
	f.authFail = false
}

func (f *form) Tick(t ticker) {
	if t == ticker1Hz && f.requestSent {
		f.authFail = f.mon.Load() == nil
	}
}

func (f *form) Draw(s Screen) {
	drawGameField(s)
	if v := f.mon.Load(); v != nil {
		m := v.(model.Monitoring)
		printHeader(s, "Usage Statistics", 0)
		s.Print(fmt.Sprintf("Players: %d\nGames: %d\nSolved: %d", m.Users, m.GamesStarted, m.GamesSolved),
			image.Point{3, 4},
			color.RGBA{0, 0xFF, 0xFF, 0xFF})
	} else {
		r := image.Rect(8, 1, len(codeInputTemplate)+4, 2)
		if f.authFail {
			s.Fill(r, color.RGBA{0xFF, 0, 0, 0xFF})
		} else if f.idx == len(f.input) {
			s.Fill(r, color.RGBA{0x80, 0x80, 0x80, 0xFF})
		}
		printHeader(s, fmt.Sprintf(codeInputTemplate, codeBox(f.input[0]), codeBox(f.input[1]), codeBox(f.input[2]), codeBox(f.input[3])), 0)
		for i := range f.dials {
			f.dials[i].Draw(s)
		}
	}
}

func (f *form) Interact(a Audio, col, row int, t time.Duration) actionResult {
	if (image.Point{col, row}).In(image.Rect(2, 1, puzzleSymX-2, 2)) {
		f.input = [4]byte{}
		f.authFail = false
		f.requestSent = false
		f.idx = 0
		return resultSwitchGame
	}
	if f.mon.Load() == nil {
		for i := range f.dials {
			if f.dials[i].Interact(col, row) {
				f.pressNextButton('0' + byte(i))
				break
			}
		}
	}
	return resultNone
}

func (f *form) pressNextButton(digit byte) {
	if f.idx >= len(f.input) {
		return
	}
	f.input[f.idx] = digit
	f.idx++
	if f.idx == len(f.input) {
		f.request(string(f.input[:]))
		f.requestSent = true
	}
}

func (f *form) Activate()        {}
func (f *form) SetLang(langCode) {}
