package main

import (
	"github.com/jroimartin/gocui"
	"github.com/sirupsen/logrus"
)

const (
	TalkList = "talk-list"
	TalkR    = 0.15
)

func main() {
	if err := Login(); err != nil {
		logrus.WithError(err).Fatal("login failed")
	}
	logrus.WithField("friend list", len(bot.FriendList)).WithField("group list", len(bot.GroupList)).Infof("debug")
	defer Shutdown()

	// enter gui mod
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		logrus.Panicf("create gui error %v", err)
	}
	defer g.Close()

	g.SetManagerFunc(layout)
}

func layout(g *gocui.Gui) error {
	return nil
}
