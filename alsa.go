package main

// #cgo LDFLAGS: -lasound
// #include <stdlib.h>
// #include <alsa/asoundlib.h>
import "C"
import (
	"errors"
	"fmt"
	"log"
	"unsafe"
)

const (
	alsaCard        = "default"
	mixerElemMaster = "Master"
)

type Mixer struct {
	mixer    *C.snd_mixer_t
	elem     *C.snd_mixer_elem_t
	channel  C.snd_mixer_selem_channel_id_t
	min, max int32
}

func alsa(e C.int) error {
	if e == 0 {
		return nil
	}
	s := C.snd_strerror(e)
	msg := C.GoString(s)
	return errors.New(msg)
}

func OpenMixer() (*Mixer, error) {
	var mix *C.snd_mixer_t
	if err := alsa(C.snd_mixer_open(&mix, 0)); err != nil {
		return nil, fmt.Errorf("alsa: opening mixer: %v", err)
	}
	defer func() {
		if mix != nil {
			if err := alsa(C.snd_mixer_close(mix)); err != nil {
				log.Printf("alsa: closing mixer: %v", err)
			}
		}
	}()

	card := C.CString(alsaCard)
	defer C.free(unsafe.Pointer(card))
	if err := alsa(C.snd_mixer_attach(mix, card)); err != nil {
		return nil, fmt.Errorf("alsa: attaching mixer master volume: %v", err)
	}

	if err := alsa(C.snd_mixer_selem_register(mix, nil, nil)); err != nil {
		return nil, fmt.Errorf("alsa: mixer register: %v", err)
	}

	if err := alsa(C.snd_mixer_load(mix)); err != nil {
		return nil, fmt.Errorf("alsa: mixer load: %v", err)
	}

	var sid *C.snd_mixer_selem_id_t
	if err := alsa(C.snd_mixer_selem_id_malloc(&sid)); err != nil {
		return nil, fmt.Errorf("alsa: allocating mixer element: %v", err)
	}

	name := C.CString(mixerElemMaster)
	defer C.free(unsafe.Pointer(name))
	C.snd_mixer_selem_id_set_name(sid, name)

	elem := C.snd_mixer_find_selem(mix, sid)
	if elem == nil {
		return nil, fmt.Errorf("alsa: cannot find mixer element")
	}

	if C.snd_mixer_selem_has_playback_volume(elem) == 0 {
		return nil, errors.New("alsa: mixer has no playback volume")
	}

	var min, max C.long
	if err := alsa(C.snd_mixer_selem_get_playback_volume_range(elem, &min, &max)); err != nil {
		return nil, fmt.Errorf("alsa: cannot get volume range: %v", err)
	}

	if C.snd_mixer_selem_has_playback_switch(elem) == 0 {
		return nil, errors.New("alsa: mixer has no mute")
	}

	m := &Mixer{
		mixer: mix,
		elem:  elem,
		min:   int32(min),
		max:   int32(max),
	}
	mix = nil
	elem = nil
	return m, nil
}

func (m *Mixer) Close() error {
	if err := alsa(C.snd_mixer_close(m.mixer)); err != nil {
		return fmt.Errorf("alsa: closing mixer: %v", err)
	}
	return nil
}

func (m *Mixer) adjust(increase bool) error {
	if m.mixer == nil || m.elem == nil {
		return errors.New("alsa: mixer is closed")
	}

	var cur C.long
	if err := alsa(C.snd_mixer_selem_get_playback_volume(m.elem, 0, &cur)); err != nil {
		return fmt.Errorf("alsa: cannot get volume: %v", err)
	}

	cur = C.long(delta5(int32(cur), m.min, m.max, increase))

	if err := alsa(C.snd_mixer_selem_set_playback_volume(m.elem, 0, cur)); err != nil {
		return fmt.Errorf("alsa: cannot set volume: %v", err)
	}

	return nil
}

func (m *Mixer) Up() error {
	return m.adjust(true)
}

func (m *Mixer) Down() error {
	return m.adjust(false)
}

func (m *Mixer) ToggleMute() error {
	if m.mixer == nil || m.elem == nil {
		return errors.New("alsa: mixer is closed")
	}

	var muted C.int
	if err := alsa(C.snd_mixer_selem_get_playback_switch(m.elem, m.channel, &muted)); err != nil {
		return fmt.Errorf("alsa: cannot get mute status: %v", err)
	}
	switch muted {
	case 0:
		muted = 1
	default:
		muted = 0
	}

	// If you run PulseAudio, it ruins your day by muting Speaker when
	// we mute Master, and never unmuting it. Directly trying to
	// unmute Speaker doesn't seem to work either, perhaps it sees our
	// change and immediately re-mutes Speaker? Weird. It's
	// reproducible in alsamixer. We explicitly don't give a *crap*
	// about PulseAudio, and didn't ask for its opinion; I recommend
	// you remove it.
	//
	// https://bugs.debian.org/cgi-bin/bugreport.cgi?bug=645063
	// https://bugs.launchpad.net/ubuntu/+source/pulseaudio/+bug/878986
	// http://askubuntu.com/questions/118675/mute-key-mutes-alsa-and-pulseaudio-but-unmutes-only-alsa
	// http://askubuntu.com/questions/339104/mute-key-mutes-master-and-headphone-speaker-alsa-channels-but-unmutes-only-ma
	// https://bugs.launchpad.net/xfce4-volumed/+bug/883485
	// http://askubuntu.com/questions/8425/how-to-temporarily-disable-pulseaudio
	//
	// Similarly, with PulseAudio, adjusting Master volume up/.down
	// causes Master, Front and PCM all adjust erratically, with
	// left/right channels becoming unbalanced. Solution: kill with fire.
	if err := alsa(C.snd_mixer_selem_set_playback_switch_all(m.elem, muted)); err != nil {
		return fmt.Errorf("alsa: cannot set mute: %v", err)
	}

	return nil
}
