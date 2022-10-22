package object

import (
	"chattweiler/internal/repository/model"
	"chattweiler/internal/vk"
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
	switch request.Command.GetMediaType() {
	case model.PictureType:
		return vk.PhotoType
	case model.AudioType:
		return vk.AudioType
	case model.VideoType:
		return vk.VideoType
	case model.DocumentType:
		return vk.DocumentType
	}

	return vk.Undefined
}
