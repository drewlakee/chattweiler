package content

import (
	"chattweiler/internal/vk"
	"github.com/SevereCloud/vksdk/v2/object"
)

type MediaAttachment struct {
	Type vk.MediaAttachmentType
	Data *object.WallWallpostAttachment
}

type AttachmentsContentCollector interface {
	CollectOne() *MediaAttachment
}
