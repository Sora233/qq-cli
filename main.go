package main

import "github.com/sirupsen/logrus"

const (
	TalkList = "talk-list"
	TalkR    = 0.15
)

func main() {
	if err := Login(); err != nil {
		logrus.WithError(err).Error("login failed")
	}

	Shutdown()
}
