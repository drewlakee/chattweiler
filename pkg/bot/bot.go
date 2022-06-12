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
	handleChatUserJoinEvent(*object.ChatUserJoinEvent)
	handleChatUserLeavingEvent(*object.ChatUserLeavingEvent)

	handleInfoCommand(*object.InfoCommand)
	handleAudioRequestCommand(*object.ContentRequestCommand)
	handlePictureRequestCommand(*object.ContentRequestCommand)

	checkMembership()

	Serve()
}
