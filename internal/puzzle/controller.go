package puzzle

import (
	"bytes"
	_ "embed"
	"encoding/hex"
	"fmt"
	"image"
	"image/color"
	"log/slog"
	"math"
	"strings"
	"sync/atomic"
	"time"
	"unicode/utf8"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/mp3"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

const (
	puzzleTileSymW = 7
	puzzleTileSymH = 3
	puzzleSymW     = 9
	puzzleSymH     = 16
	puzzleSymX     = 31
	puzzleSymY     = 16
	puzzlePixX     = puzzleSymX * puzzleSymW
	puzzlePixY     = puzzleSymY * puzzleSymH

	logLineH = 18
)

var (
	//go:embed assets/fonts/font9x16.ttf
	DosFont []byte
	//go:embed assets/audio/click.mp3
	Click_mp3 []byte
	//go:embed assets/text/game_field.txt
	playFieldTemplate string
	//go:embed assets/text/splash_screen.txt
	splashScreenTemplate string
	//go:embed assets/text/gopher_cyan.txt
	gopherCyan string
	//go:embed assets/text/gopher_white.txt
	gopherWhite string
	//go:embed assets/text/gopher_beige.txt
	gopherBeige string
	//go:embed assets/text/gopher_gray.txt
	gopherGray string

	defaultBgColor = color.Black
)

type screen int

const (
	screenGame screen = iota
	screenForm
	screenSplash
	screenDebug
)

type ticker int

const (
	ticker1Hz ticker = iota
	ticker2Hz
	ticker4Hz
	ticker10Hz
	ticker25Hz
)

type actionResult int

const (
	resultNone actionResult = iota
	resultSwitchGame
	resultSwitchForm
	resultSwitchDebug
)

type Audio interface {
	PlaySound()
}

type Screen interface {
	Fill(rect image.Rectangle, clr color.Color)
	Print(txt string, coord image.Point, clr color.Color)
	PrintDebug(msg string, line int)
}

type Handler interface {
	Draw(Screen)
	Tick(ticker)
	Interact(Audio, int, int, time.Duration) actionResult
	SetLang(langCode)
	Activate()
}

type Controller struct {
	screen  image.Rectangle
	bgColor color.Color
	scr     *ebiten.Image
	font    *text.GoTextFaceSource

	activeState  atomic.Bool
	activeScreen screen
	screens      map[screen]Handler

	debugFn func(string)

	btnPressed        time.Time
	touchTapped       map[ebiten.TouchID]time.Time
	OnGameStart       func()
	OnGameSolve       func(int)
	InfoRequest       func()
	UserStatsRequest  func()
	MonitoringRequest func(string)
	UrlOpener         func(string)

	audioCtx *audio.Context
	player   *audio.Player
}

func Init(init ...func(*Controller)) error {
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetWindowSize(640, 480)
	ebiten.SetWindowTitle(l10nGameTitle(langCodeEn))
	c, err := newController()
	if err != nil {
		return err
	}
	for i := range init {
		init[i](c)
	}
	c.screens[screenGame] = newGame(c.OnGameStart, c.OnGameSolve, c.UserStatsRequest)
	c.screens[screenForm] = newForm(c.MonitoringRequest)
	c.screens[screenSplash] = newSplash(c.UrlOpener)
	c.screens[screenDebug], c.debugFn = newDebugOverlay()

	c.SetLangCode(langCodeRu)

	return ebiten.RunGame(c)
}

func newController() (*Controller, error) {
	tfs, err := text.NewGoTextFaceSource(bytes.NewReader(DosFont))
	if err != nil {
		return nil, err
	}

	audioCtx := audio.NewContext(44100)
	s, err := mp3.DecodeWithoutResampling(bytes.NewReader(Click_mp3))
	if err != nil {
		return nil, err
	}
	player, err := audioCtx.NewPlayer(s)
	if err != nil {
		return nil, err
	}

	p := &Controller{
		scr:          ebiten.NewImage(puzzlePixX, puzzlePixY),
		font:         tfs,
		activeScreen: screenSplash,
		screens:      make(map[screen]Handler),
		touchTapped:  make(map[ebiten.TouchID]time.Time),
		player:       player,
		audioCtx:     audioCtx,
		activeState:  atomic.Bool{},
	}
	p.OnGameStart = func() {}
	p.OnGameSolve = func(int) {}
	p.InfoRequest = func() {}
	p.UserStatsRequest = func() {}
	p.MonitoringRequest = func(string) {}
	p.UrlOpener = func(url string) {
		if err := openURL(url); err != nil {
			p.Debug("url open error: %v", err)
		}
	}
	go p.tick()
	return p, nil
}

func (c *Controller) tick() {
	t1Hz := time.NewTicker(time.Second)
	t2Hz := time.NewTicker(time.Millisecond * 500)
	t4Hz := time.NewTicker(time.Millisecond * 250)
	t10Hz := time.NewTicker(time.Millisecond * 100)
	t25Hz := time.NewTicker(time.Millisecond * 40)
	for {
		select {
		case <-t1Hz.C:
			c.TickEvent(ticker1Hz)
		case <-t2Hz.C:
			c.TickEvent(ticker2Hz)
		case <-t4Hz.C:
			c.TickEvent(ticker4Hz)
		case <-t10Hz.C:
			c.TickEvent(ticker10Hz)
		case <-t25Hz.C:
			c.TickEvent(ticker25Hz)
		}
	}
}

func (c *Controller) SetActive(active bool) {
	c.activeState.Store(active)
}

func (c *Controller) TickEvent(t ticker) {
	c.screens[c.activeScreen].Tick(t)
}

func (c *Controller) Update() error {
	if !c.activeState.Load() {
		return nil
	}
	switch {
	case inpututil.IsMouseButtonJustPressed(ebiten.MouseButton0):
		c.btnPressed = time.Now()
	case inpututil.IsMouseButtonJustReleased(ebiten.MouseButton0):
		x, y := ebiten.CursorPosition()
		c.interact(x, y, time.Since(c.btnPressed))
		c.btnPressed = time.Time{}
	}

	for _, id := range inpututil.AppendJustPressedTouchIDs(make([]ebiten.TouchID, 0)) {
		c.touchTapped[id] = time.Now()
	}

	for id, v := range c.touchTapped {
		if v.IsZero() || v.Add(time.Second*5).Before(time.Now()) {
			delete(c.touchTapped, id)
		}
	}

	for _, id := range inpututil.AppendJustReleasedTouchIDs(make([]ebiten.TouchID, 0)) {
		if v, ok := c.touchTapped[id]; ok {
			x, y := inpututil.TouchPositionInPreviousTick(id)
			c.interact(x, y, time.Since(v))
		}
		delete(c.touchTapped, id)
	}

	return nil
}

func (c *Controller) Draw(screen *ebiten.Image) {
	if !c.activeState.Load() {
		return
	}
	screen.Fill(c.backgroundColor())

	c.drawMatrixScreen(c.calcBounds(screen))
}

func (c *Controller) PrintDebug(msg string, line int) {
	ebitenutil.DebugPrintAt(c.scr, msg, 0, c.scr.Bounds().Dy()-logLineH*(line+1)-12)
}

func (c *Controller) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return outsideWidth, outsideHeight
}

func (c *Controller) PlaySound() {
	c.player.SetVolume(0.2)
	c.player.SetPosition(0)
	c.player.Play()
}

func (c *Controller) interact(x, y int, t time.Duration) {
	col := int(math.Floor(float64(x-c.screen.Min.X) / (float64(puzzleSymW) * (float64(c.screen.Bounds().Dx()) / float64(c.scr.Bounds().Dx())))))
	row := int(math.Floor(float64(y-c.screen.Min.Y) / (float64(puzzleSymH) * (float64(c.screen.Bounds().Dy()) / float64(c.scr.Bounds().Dy())))))

	switch c.screens[c.activeScreen].Interact(c, col, row, t) {
	case resultSwitchDebug:
		c.screens[screenDebug].Activate()
		fallthrough
	case resultSwitchGame:
		c.switchScreen(screenGame)
	case resultSwitchForm:
		c.switchScreen(screenForm)
	default:
		// NOOP
	}
}

func (c *Controller) switchScreen(s screen) {
	c.btnPressed = time.Time{}
	c.activeScreen = s
	c.screens[c.activeScreen].Activate()
}

func (c *Controller) Debug(format string, a ...any) {
	msg := fmt.Sprintf(format, a...)
	slog.Default().Debug(msg)
	c.debugFn(msg)
}

func (c *Controller) calcBounds(screen *ebiten.Image) *ebiten.Image {
	bounds := screen.Bounds()
	var rec = &image.Rectangle{Min: image.Point{}, Max: image.Point{bounds.Dx(), bounds.Dy()}}

	kefX := float64(puzzlePixX) / float64(puzzlePixY)
	kefY := float64(puzzlePixY) / float64(puzzlePixX)

	screenX := math.Floor(kefX * float64(bounds.Dy()))
	screenY := math.Floor(kefY * float64(bounds.Dx()))

	if int(screenX) > bounds.Dx() {
		screenX = kefX * screenY
	}
	if int(screenY) > bounds.Dy() {
		screenY = kefY * screenX
	}

	rec.Min.X = (bounds.Dx() - int(screenX)) / 2
	rec.Max.X = bounds.Dx() - rec.Min.X
	rec.Min.Y = (bounds.Dy() - int(screenY)) / 2
	rec.Max.Y = bounds.Dy() - rec.Min.Y

	c.screen = *rec

	return screen.SubImage(*rec).(*ebiten.Image)
}

func (c *Controller) drawMatrixScreen(scr *ebiten.Image) {
	c.scr.Clear()

	c.screens[c.activeScreen].Draw(c)
	c.screens[screenDebug].Draw(c) // appears as overlay when active

	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Scale(float64(scr.Bounds().Dx())/float64(c.scr.Bounds().Dx()), float64(scr.Bounds().Dy())/float64(c.scr.Bounds().Dy()))
	opts.GeoM.Translate(float64(scr.Bounds().Min.X), float64(scr.Bounds().Min.Y))
	scr.DrawImage(c.scr, opts)
}

func drawGameField(s Screen) {
	s.Fill(image.Rect(0, 0, puzzleSymX, puzzleSymY), // header and all field
		nil) // background: outside background (screen)

	s.Fill(image.Rect(1, 0, puzzleSymX-1, 2), // header only
		color.RGBA{0x40, 0xA0, 0x40, 0xFF}) // background: yellow

	s.Fill(image.Rect(1, 2, puzzleSymX-1, puzzleSymY), // play field
		color.RGBA{0, 0x00, 0xA0, 0xFF}) // background: ALWAYS same as tile

	s.Print(playFieldTemplate, // border
		image.Point{0, 0},
		color.RGBA{0xFF, 0xFF, 0xFF, 0xFF}) // text: white
}

func (c *Controller) backgroundColor() color.Color {
	if c.bgColor == nil {
		return defaultBgColor
	}
	return c.bgColor
}

func (c *Controller) Fill(rect image.Rectangle, color color.Color) {
	if color == nil {
		color = c.backgroundColor()
	}
	(c.scr.SubImage(image.Rect(
		rect.Min.X*puzzleSymW,
		rect.Min.Y*puzzleSymH,
		int(math.Min(float64(rect.Max.X), float64(puzzleSymX)))*puzzleSymW,
		int(math.Min(float64(rect.Max.Y), float64(puzzleSymY)))*puzzleSymH))).(*ebiten.Image).Fill(color)
}

func (c *Controller) Print(txt string, coord image.Point, clr color.Color) {
	op := &text.DrawOptions{}
	op.GeoM.Translate(float64(coord.X*puzzleSymW), float64(coord.Y*puzzleSymH))
	op.ColorScale.ScaleWithColor(clr)
	op.LineSpacing = float64(puzzleSymH)
	op.PrimaryAlign = text.AlignStart
	op.SecondaryAlign = text.AlignStart
	text.Draw(c.scr, txt, &text.GoTextFace{
		Source: c.font,
		Size:   float64(puzzleSymH),
	}, op)
}

func printHeader(s Screen, text string, align int) {
	var p image.Point
	l := utf8.RuneCountInString(text)
	switch {
	case align < 0:
		p = image.Point{2, 1}
	case align > 0:
		p = image.Point{puzzleSymX - 2 - l, 1}
	default:
		p = image.Point{(puzzleSymX - l) / 2, 1}
	}
	s.Print(text, p, color.White)
}

func (c *Controller) OnLoad(loadSec float64, bgColor, langCode string) {
	c.InfoRequest()
	c.SetLangCode(langCode)
	c.SetBgColor(bgColor)
}

func (c *Controller) SetLangCode(lc string) {
	for _, s := range c.screens {
		s.SetLang(l10nCode(lc))
	}
}

func (c *Controller) SetBgColor(clr string) {
	colorStr, err := normalize(clr)
	if err != nil {
		c.Debug(fmt.Sprintf("ERROR: %s", err))
	} else {
		b, err := hex.DecodeString(colorStr)
		switch {
		case err != nil:
			c.Debug(fmt.Sprintf("ERROR: %s", err))
		case c.bgColor == nil:
			fallthrough
		default:
			c.bgColor = color.RGBA{b[0], b[1], b[2], b[3]}
		}
	}
}

func normalize(colorStr string) (string, error) {
	offset := 0
	if strings.HasPrefix(colorStr, "#") {
		offset = 1
	}
	b := colorStr[offset:]
	if len(b) == 8 {
		return b, nil
	}
	if len(b) != 6 {
		return "", fmt.Errorf("normalize: not enough symbols: '%s'", colorStr)
	}
	return b + "FF", nil
}
