// Package bot provides the bot interface
package bot

import (
	"chattweiler/pkg/bot/object"
)

var logPackage = "bot"

type Bot interface {
	Serve()

	handleChatUserJoinEvent(*object.ChatEvent)
	handleChatUserLeavingEvent(*object.ChatEvent)

	handleInfoCommand(*object.ChatEvent)
	handleContentRequestCommand(command *object.ContentRequestCommand)
}
