// Command x11-media-keys is a user daemon that reacts to X11 media
// keys, such as volume up/down, and makes the desired changes.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/BurntSushi/xgb/randr"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/keybind"
	"github.com/BurntSushi/xgbutil/xevent"
)

var prog = filepath.Base(os.Args[0])

func usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", prog)
	fmt.Fprintf(os.Stderr, "  %s\n", prog)
	fmt.Fprintf(os.Stderr, "(the command takes no options)\n")
}

type binding struct {
	key string
	fn  keybind.KeyPressFun
}

func bind(Xu *xgbutil.XUtil, bindings ...binding) error {
	for _, k := range bindings {
		if err := k.fn.Connect(Xu, Xu.RootWin(), k.key, true); err != nil {
			return err
		}
	}
	return nil
}

func run() error {
	Xu, err := xgbutil.NewConn()
	if err != nil {
		return err
	}
	defer Xu.Conn().Close()
	keybind.Initialize(Xu)
	if err := randr.Init(Xu.Conn()); err != nil {
		return err
	}

	audio, err := newAudio(Xu)
	if err != nil {
		return err
	}
	defer audio.Close()

	brightness, err := newBrightness(Xu)
	if err != nil {
		return err
	}
	defer brightness.Close()

	xevent.Main(Xu)
	return nil
}

func main() {
	log.SetFlags(0)
	log.SetPrefix(prog + ": ")

	flag.Usage = usage
	flag.Parse()
	if flag.NArg() != 0 {
		flag.Usage()
		os.Exit(2)
	}

	if err := run(); err != nil {
		log.Fatal(err)
	}
}
