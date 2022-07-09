// Package content provides services for content fetching
package content

import (
	"github.com/SevereCloud/vksdk/v2/object"
)

type AttachmentsContentCollector interface {
	CollectOne() object.WallWallpostAttachment
}
