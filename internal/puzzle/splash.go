package puzzle

import (
	"15-puzzle/internal/model"
	"image"
	"image/color"
	"sync/atomic"
	"time"
)

const (
	gopherX = 8
	gopherY = 1
)

type splash struct {
	langCode  langCode
	urlOpener func(string)
	gopherDx  int
	info      atomic.Value
}

func newSplash(urlOpener func(string)) *splash {
	f := &splash{langCode: langCodeEn, urlOpener: urlOpener, gopherDx: 22}
	return f
}

func (sp *splash) SetLang(lc langCode) {
	sp.langCode = lc
}

func (sp *splash) Tick(t ticker) {
	if t == ticker25Hz && sp.gopherDx > 0 {
		sp.gopherDx--
	}
}

func (sp *splash) ApiInfoHandler(i model.Info) {
	sp.info.Store(i)
}

func (sp *splash) Interact(a Audio, col, row int, t time.Duration) actionResult {
	if sp.gopherDx > 0 {
		return resultNone
	}
	if info := sp.info.Load(); info != nil && row == puzzleSymY-2 && col < puzzleSymX-2 && col > puzzleSymX-9 {
		sp.urlOpener(info.(model.Info).ProjectLink)
		return resultNone
	}
	if t > time.Second*10 {
		return resultSwitchDebug
	}
	return resultSwitchGame
}

func (sp *splash) Draw(s Screen) {
	s.Fill(image.Rect(0, 0, puzzleSymX, puzzleSymY), color.RGBA{0x0A, 0x23, 0x4E, 0xFF})
	sp.drawGopher(s)
	s.Print(splashScreenTemplate, image.Point{0, 0}, color.RGBA{0x80, 0x80, 0x80, 0xFF})
	s.Print(`"`+l10nGameTitle(sp.langCode)+`"`, image.Point{1, 1}, color.White)
	sp.drawProjectLink(s)
}

func (sp *splash) drawGopher(s Screen) {
	x := gopherX + sp.gopherDx
	s.Print(gopherCyan, image.Point{x, gopherY}, color.RGBA{0x00, 0xFF, 0xFF, 0xFF})
	s.Print(gopherWhite, image.Point{x, gopherY}, color.White)
	s.Print(gopherBeige, image.Point{x, gopherY}, color.RGBA{0xF4, 0xB3, 0x6D, 0xFF})
	s.Print(gopherGray, image.Point{x, gopherY}, color.RGBA{0x80, 0x80, 0x80, 0xFF})
	s.Fill(image.Rect(puzzleSymX-1, 0, puzzleSymX, puzzleSymY), color.RGBA{0x0A, 0x23, 0x4E, 0xFF})
}

func (sp *splash) drawProjectLink(s Screen) {
	if sp.info.Load() != nil {
		s.Print("______", image.Point{puzzleSymX - 8, puzzleSymY - 2}, color.RGBA{0x00, 0xFF, 0xFF, 0xFF})
	}
	s.Print("GitHub", image.Point{puzzleSymX - 8, puzzleSymY - 2}, color.RGBA{0xFB, 0xE6, 0x8E, 0xFF})
}

func (sp *splash) Activate() {}
