package main

import (
	"errors"
	"fmt"
	"log"
	"unsafe"

	"github.com/BurntSushi/xgb/randr"
	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/xevent"
)

type brightness struct {
	prop xproto.Atom
}

func (b *brightness) adjust(Xu *xgbutil.XUtil, increase bool) error {
	X := Xu.Conn()
	root := xproto.Setup(X).DefaultScreen(X).Root
	screens, err := randr.GetScreenResources(X, root).Reply()
	if err != nil {
		return fmt.Errorf("getting screen: %v", err)
	}

	for _, output := range screens.Outputs {
		query, err := randr.QueryOutputProperty(X, output, b.prop).Reply()
		if err != nil {
			if _, ok := err.(xproto.NameError); ok {
				// this output has no backlight
				continue
			}
			return fmt.Errorf("query backlight: %v", err)
		}
		if !query.Range {
			return errors.New("backlight brightness range not specified")
		}
		if len(query.ValidValues) != 2 {
			return fmt.Errorf("expected min and max, got: %v", query.ValidValues)
		}
		min, max := query.ValidValues[0], query.ValidValues[1]
		// log.Printf("backlight range: %d .. %d", min, max)

		get, err := randr.GetOutputProperty(X, output, b.prop, xproto.AtomNone, 0, 4, false, false).Reply()
		if err != nil {
			return fmt.Errorf("get backlight property: %v", err)
		}
		if get.Type != xproto.AtomInteger ||
			get.NumItems != 1 ||
			get.Format != 32 {
			return fmt.Errorf("backlight property value looks wrong")
		}
		old := *(*int32)(unsafe.Pointer(&get.Data[0]))
		// log.Printf("backlight data: %d", old)

		bri := delta5(old, min, max, increase)

		data := (*[4]byte)(unsafe.Pointer(&bri))[:]
		if err := randr.ChangeOutputPropertyChecked(X, output, b.prop, xproto.AtomInteger, 32, xproto.PropModeReplace, 1, data).Check(); err != nil {
			return err
		}
	}
	return nil
}

func (b *brightness) up(X *xgbutil.XUtil, ev xevent.KeyPressEvent) {
	if err := b.adjust(X, true); err != nil {
		log.Fatalf("error adjusting brightness: %v", err)
	}
}

func (b *brightness) down(X *xgbutil.XUtil, ev xevent.KeyPressEvent) {
	if err := b.adjust(X, false); err != nil {
		log.Fatalf("error adjusting brightness: %v", err)
	}
}

func (b *brightness) Close() {
	// nothing
}

func newBrightness(Xu *xgbutil.XUtil) (*brightness, error) {
	const atomName = "Backlight"
	atomReply, err := xproto.InternAtom(Xu.Conn(), true, uint16(len(atomName)), atomName).Reply()
	if err != nil {
		return nil, fmt.Errorf("no backlight: %v", err)
	}

	b := &brightness{
		prop: atomReply.Atom,
	}

	if err := bind(Xu,
		binding{"XF86MonBrightnessUp", b.up},
		binding{"XF86MonBrightnessDown", b.down},
	); err != nil {
		return nil, err
	}

	return b, nil
}
