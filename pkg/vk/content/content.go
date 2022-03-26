package content

import "github.com/SevereCloud/vksdk/v2/object"

type AttachmentsType string

const (
	Audio AttachmentsType = "audio"
	Photo AttachmentsType = "photo"
)

type AttachmentsContentCollector interface {
	CollectOne() object.WallWallpostAttachment
}
