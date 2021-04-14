package main

import (
	"fmt"
	"github.com/jroimartin/gocui"
	"github.com/sirupsen/logrus"
	"time"
)

const (
	TalkList = "talk-list"
	TalkR    = 0.15
)

var talkEntries []string
var talkChan = make(chan string)

func main() {
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		logrus.Panicf("create gui error %v", err)
	}
	defer g.Close()
	g.SetManagerFunc(layout)

	if err := keybinding(g); err != nil {
		logrus.Panicf("key binding error %v", err)
	}

	go func() {
		for range time.Tick(time.Second * 3) {
			talkEntries = append(talkEntries, "a")
			g.Update(func(gui *gocui.Gui) error {
				return nil
			})
		}
	}()

	if err := g.MainLoop(); err != nil {
		if err != gocui.ErrQuit {
			logrus.Panicf("main loop error %v", err)
		}
	}
}

func keybinding(g *gocui.Gui) error {
	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone,
		func(g *gocui.Gui, v *gocui.View) error {
			return gocui.ErrQuit
		}); err != nil {
		return err
	}
	if err := g.SetKeybinding("", gocui.KeySpace, gocui.ModNone,
		func(gui *gocui.Gui, view *gocui.View) error {
			v, err := gui.View(TalkList)
			if err != nil {
				return err
			}
			v.Frame = !v.Frame
			return nil
		}); err != nil {
		return err
	}
	return nil
}

func createTalkList(g *gocui.Gui) (*gocui.View, error) {
	x, y := g.Size()
	v, err := g.SetView(TalkList, 0, 0, int(float64(x-1)*TalkR), y-1)
	if err == gocui.ErrUnknownView {
		// new view
		v.Title = "Talk"
		v.Editable = false
	} else if err != nil {
		return nil, err
	} else {
		// update view
		x, y = v.Size()
		for index, str := range talkEntries {
			v, _ = g.SetView(fmt.Sprintf("%v-%v", TalkList, str))
			v.Wrap = true
		}
	}
	return v, nil
}

func layout(g *gocui.Gui) error {
	_, err := createTalkList(g)
	if err != nil {
		return err
	}
	return nil
}
