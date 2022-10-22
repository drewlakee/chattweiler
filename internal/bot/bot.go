package bot

import (
	"chattweiler/internal/bot/object"
)

var logPackage = "bot"

type Bot interface {
	Serve()

	handleChatUserJoinEvent(*object.ChatEvent)
	handleChatUserLeavingEvent(*object.ChatEvent)

	handleInfoCommand(*object.ChatEvent)
	handleContentRequestCommand(command *object.ContentRequestCommand)
}
