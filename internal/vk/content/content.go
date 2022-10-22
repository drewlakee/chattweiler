package content

import (
	"github.com/SevereCloud/vksdk/v2/object"
)

type AttachmentsContentCollector interface {
	CollectOne() object.WallWallpostAttachment
}
