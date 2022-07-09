// Package object provides structures which are used by bot
package object

import (
	"chattweiler/pkg/repository/model"
	"chattweiler/pkg/vk"
)

type ChatEvent struct {
	UserID string
	PeerID int
}

type ContentRequestCommand struct {
	Command *model.ContentCommand
	Event   *ChatEvent
}

func (request *ContentRequestCommand) GetAttachmentsType() vk.AttachmentsType {
	switch request.Command.MediaContentType {
	case model.PictureType:
		return vk.PhotoType
	case model.AudioType:
		return vk.AudioType
	}

	return vk.Undefined
}
