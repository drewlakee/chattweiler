package content

import (
	"github.com/SevereCloud/vksdk/v2/object"
	"github.com/sirupsen/logrus"
)

var packageLogFields = logrus.Fields{
	"package": "content",
}

type AttachmentsType string

const (
	AudioType AttachmentsType = "audio"
	PhotoType AttachmentsType = "photo"
)

type AttachmentsContentCollector interface {
	CollectOne() object.WallWallpostAttachment
}
