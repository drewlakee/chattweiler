package mapper

import (
	"chattweiler/internal/repository/model"
	wrapper "github.com/SevereCloud/vksdk/v2/longpoll-user/v3"

	"chattweiler/internal/bot/object"
	"strconv"
)

func NewChatEventFromFromChatInfoChange(event wrapper.ChatInfoChange) *object.ChatEvent {
	return &object.ChatEvent{
		UserID: strconv.Itoa(event.Info),
		PeerID: event.PeerID,
	}
}

func NewChatEventFromNewMessage(message wrapper.NewMessage) *object.ChatEvent {
	return &object.ChatEvent{
		PeerID: message.PeerID,
	}
}

func NewContentCommandRequest(command *model.Command, message wrapper.NewMessage) *object.ContentRequestCommand {
	return &object.ContentRequestCommand{
		Command: command,
		Event: &object.ChatEvent{
			UserID: message.AdditionalData.From,
			PeerID: message.PeerID,
		},
	}
}
