// Package provides map functions from API objects
// to application objects
package mapper

import (
	wrapper "github.com/SevereCloud/vksdk/v2/longpoll-user/v3"

	"chattweiler/pkg/bot/object"
	"chattweiler/pkg/vk"
	"strconv"
)

func NewChatUserJoinEventFromChatInfoChange(event wrapper.ChatInfoChange) *object.ChatUserJoinEvent {
	return &object.ChatUserJoinEvent{
		UserID: strconv.Itoa(event.Info),
		PeerID: event.PeerID,
	}
}

func NewChatUserLeavingEventFromChatInfoChange(event wrapper.ChatInfoChange) *object.ChatUserLeavingEvent {
	return &object.ChatUserLeavingEvent{
		UserID: strconv.Itoa(event.Info),
		PeerID: event.PeerID,
	}
}

func NewInfoCommandFromNewMessage(message wrapper.NewMessage) *object.InfoCommand {
	return &object.InfoCommand{
		PeerID: message.PeerID,
	}
}

func NewContentRequestCommandFromNewMessage(contentType vk.AttachmentsType, message wrapper.NewMessage) *object.ContentRequestCommand {
	return &object.ContentRequestCommand{
		Type:   contentType,
		UserID: message.AdditionalData.From,
		PeerID: message.PeerID,
	}
}
