package puzzle

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"strconv"
	"time"
)

type tileFormatter string

func (t tileFormatter) Format(c string) string {
	return fmt.Sprintf(string(t), c[:int(math.Min(float64(len(c)), 2))])
}

const tileTemplate tileFormatter = `╔════╗
║ %2s ║
╚════╝`

type button struct {
	title func() string
	x, y  func() int
}

func (t *button) Interact(x, y int) bool {
	if (image.Point{x, y}).In(image.Rect(t.x(), t.y(), t.x()+puzzleTileSymW, t.y()+puzzleTileSymH)) {
		t.onTrigger()
		return true
	}
	return false
}

func (t *button) Title() string {
	return t.title()
}

func (t *button) Draw(s Screen) {
	s.Print(tileTemplate.Format(t.Title()),
		image.Point{t.x(), t.y()},
		color.RGBA{0, 0xFF, 0xFF, 0xFF},
	)
}

func (t *button) onTrigger() {
}

func NewButton(title func() string, x, y func() int) *button {
	return &button{
		title: title,
		x:     x,
		y:     y,
	}
}

type tile struct {
	button
	num int
	pos int

	moving bool
	dx, dy float64
}

func NewTile(num, pos int) *tile {
	t := &tile{
		num: num,
		pos: pos,
	}
	t.button = *NewButton(t.title, t.X, t.Y)
	return t
}

func (t *tile) Title() string {
	if t.num == 0 {
		return "Go"
	}
	return strconv.Itoa(t.num)
}

func (t *tile) CanInteract(x, y int) bool {
	return !t.moving && t.button.Interact(x, y)
}

func (t *tile) MoveUp(finish func()) {
	t.shift(&t.dy, -1, finish)
}

func (t *tile) MoveDown(finish func()) {
	t.shift(&t.dy, 1, finish)
}

func (t *tile) MoveLeft(finish func()) {
	t.shift(&t.dx, -1, finish)
}

func (t *tile) MoveRight(finish func()) {
	t.shift(&t.dx, 1, finish)
}

func (t *tile) shift(src *float64, target float64, finish func()) {
	if t.moving {
		return
	}
	t.moving = true
	*src = 0
	go t.shifter(src, target, func() { finish(); t.moving = false })
}

func (t *tile) shifter(src *float64, target float64, finish func()) {
	step := target / 10
	ticker := time.NewTicker(time.Millisecond * 10)
	for {
		<-ticker.C
		*src += step
		if math.Abs(*src) >= math.Abs(target) {
			*src = 0
			break
		}
	}
	finish()
}

func (t *tile) Col() int {
	return t.pos % 4
}

func (t *tile) Row() int {
	return t.pos / 4
}

func (t *tile) X() int {
	return t.Col()*puzzleTileSymW + int(t.dx*puzzleTileSymW) + 2
}

func (t *tile) Y() int {
	return t.Row()*puzzleTileSymH + int(t.dy*puzzleTileSymH) + 3
}
