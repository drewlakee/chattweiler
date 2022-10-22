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

func (request *ContentRequestCommand) GetAttachmentsTypes() []vk.MediaAttachmentType {
	var types []vk.MediaAttachmentType
	for _, t := range request.Command.GetMediaTypes() {
		switch t {
		case model.PictureType:
			types = append(types, vk.PhotoType)
		case model.AudioType:
			types = append(types, vk.AudioType)
		case model.VideoType:
			types = append(types, vk.VideoType)
		case model.DocumentType:
			types = append(types, vk.DocumentType)
		}
	}

	return types
}
