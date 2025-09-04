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

type stats struct {
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

func newStats(request func(string)) *stats {
	col := func(i int) int { return (i%3)*puzzleTileSymW + 5 }
	row := func(i int) int { return (i/3)*puzzleTileSymH + 3 }
	f := &stats{request: request}
	for i := 1; i < dials; i++ {
		f.dials[i] = NewButton(stringFn(strconv.Itoa(i)), intFn(col(i-1)), intFn(row(i-1)))
	}
	f.dials[0] = NewButton(stringFn("0"), intFn(col(10)), intFn(row(10)))
	return f
}

func (st *stats) ApiMonitoringHandler(m model.Monitoring) {
	st.mon.Store(m)
	st.authFail = false
}

func (st *stats) Tick(t ticker) {
	if t == ticker1Hz && st.requestSent {
		st.authFail = st.mon.Load() == nil
	}
}

func (st *stats) Draw(s Screen) {
	drawGameField(s)
	if v := st.mon.Load(); v != nil {
		m := v.(model.Monitoring)
		printHeader(s, "Usage Statistics", 0)
		s.Print(fmt.Sprintf("Players: %d\nGames: %d\nSolved: %d", m.Users, m.GamesStarted, m.GamesSolved),
			image.Point{3, 4},
			color.RGBA{0, 0xFF, 0xFF, 0xFF})
	} else {
		r := image.Rect(8, 1, len(codeInputTemplate)+4, 2)
		if st.authFail {
			s.Fill(r, color.RGBA{0xFF, 0, 0, 0xFF})
		} else if st.idx == len(st.input) {
			s.Fill(r, color.RGBA{0x80, 0x80, 0x80, 0xFF})
		}
		printHeader(s, fmt.Sprintf(codeInputTemplate, codeBox(st.input[0]), codeBox(st.input[1]), codeBox(st.input[2]), codeBox(st.input[3])), 0)
		for i := range st.dials {
			st.dials[i].Draw(s)
		}
	}
}

func (st *stats) Interact(a Audio, col, row int, t time.Duration) actionResult {
	if (image.Point{col, row}).In(image.Rect(2, 1, puzzleSymX-2, 2)) {
		st.input = [4]byte{}
		st.authFail = false
		st.requestSent = false
		st.idx = 0
		return resultSwitchGame
	}
	if st.mon.Load() == nil {
		for i := range st.dials {
			if st.dials[i].Interact(col, row) {
				st.pressNextButton('0' + byte(i))
				break
			}
		}
	}
	return resultNone
}

func (st *stats) pressNextButton(digit byte) {
	if st.idx >= len(st.input) {
		return
	}
	st.input[st.idx] = digit
	st.idx++
	if st.idx == len(st.input) {
		st.request(string(st.input[:]))
		st.requestSent = true
	}
}

func (st *stats) Activate()        {}
func (st *stats) SetLang(langCode) {}
