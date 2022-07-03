// Package provides map functions from API objects
// to application objects
package mapper

import (
	"github.com/SevereCloud/vksdk/v2/events"
	wrapper "github.com/SevereCloud/vksdk/v2/longpoll-user/v3"

	"chattweiler/pkg/bot/object"
	"chattweiler/pkg/vk"
	"strconv"
)

func NewChatEventFromFromChatInfoChange(event wrapper.ChatInfoChange) *object.ChatEvent {
	return &object.ChatEvent{
		UserID: strconv.Itoa(event.Info),
		PeerID: event.PeerID,
	}
}

func NewChatEventFromMessageNewObject(event events.MessageNewObject) *object.ChatEvent {
	return &object.ChatEvent{
		UserID: strconv.Itoa(event.Message.Action.MemberID),
		PeerID: event.Message.PeerID,
	}
}

func NewChatEventFromNewMessage(message wrapper.NewMessage) *object.ChatEvent {
	return &object.ChatEvent{
		PeerID: message.PeerID,
	}
}

func NewContentRequestCommandFromNewMessage(contentType vk.AttachmentsType, message wrapper.NewMessage) *object.ContentRequestCommand {
	return &object.ContentRequestCommand{
		Type: contentType,
		RequestEvent: &object.ChatEvent{
			UserID: message.AdditionalData.From,
			PeerID: message.PeerID,
		},
	}
}
