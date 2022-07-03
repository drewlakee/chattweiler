// Package provides a bot interface for the application,
// and it's implementations
package bot

import (
	"chattweiler/pkg/bot/object"

	"github.com/sirupsen/logrus"
)

var packageLogFields = logrus.Fields{
	"package": "bot",
}

type Bot interface {
	Serve()

	handleChatUserJoinEvent(*object.ChatEvent)
	handleChatUserLeavingEvent(*object.ChatEvent)

	handleInfoCommand(*object.ChatEvent)
	handleAudioRequestCommand(*object.ContentRequestCommand)
	handlePictureRequestCommand(*object.ContentRequestCommand)
}
