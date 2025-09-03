package puzzle

import "time"

type debug struct {
	idx      int
	debugMsg []string
	active   bool
}

func newDebugOverlay() (*debug, func(string)) {
	d := &debug{
		debugMsg: make([]string, 10),
	}
	return d, d.AddMessage
}

func (d *debug) AddMessage(msg string) {
	d.debugMsg[d.idx] = msg
	d.idx = (d.idx + 1) % len(d.debugMsg)
}

func (d *debug) Draw(s Screen) {
	if !d.active || len(d.debugMsg) == 0 {
		return
	}
	for i := range d.debugMsg {
		s.PrintDebug(d.debugMsg[i], i)
	}
}

func (d *debug) Activate()                                            { d.active = true }
func (d *debug) Interact(Audio, int, int, time.Duration) actionResult { return resultNone }
func (d *debug) Tick(ticker)                                          {}
func (d *debug) SetLang(langCode)                                     {}
