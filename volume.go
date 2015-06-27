package main

import (
	"log"

	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/keybind"
	"github.com/BurntSushi/xgbutil/xevent"
)

type audio struct {
	mixer *Mixer
}

func (a *audio) up(X *xgbutil.XUtil, ev xevent.KeyPressEvent) {
	if err := a.mixer.Up(); err != nil {
		log.Print(err)
	}
}

func (a *audio) down(X *xgbutil.XUtil, ev xevent.KeyPressEvent) {
	if err := a.mixer.Down(); err != nil {
		log.Print(err)
	}
}

func (a *audio) mute(X *xgbutil.XUtil, ev xevent.KeyPressEvent) {
	if err := a.mixer.ToggleMute(); err != nil {
		log.Print(err)
	}
}

func (a *audio) Close() {
	_ = a.mixer.Close()
	a.mixer = nil
}

func newAudio(Xu *xgbutil.XUtil) (*audio, error) {
	m, err := OpenMixer()
	if err != nil {
		return nil, err
	}
	defer func() {
		// set m to nil on final success
		if m != nil {
			_ = m.Close()
		}
	}()

	a := &audio{
		mixer: m,
	}

	for _, k := range []struct {
		key string
		fn  keybind.KeyPressFun
	}{
		{"XF86AudioRaiseVolume", a.up},
		{"XF86AudioLowerVolume", a.down},
		{"XF86AudioMute", a.mute},
	} {
		if err := k.fn.Connect(Xu, Xu.RootWin(), k.key, true); err != nil {
			return nil, err
		}
	}

	m = nil
	return a, nil
}
