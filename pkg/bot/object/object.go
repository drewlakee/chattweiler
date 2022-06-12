// Package provides object sturctures which bot uses
// to operate
package object

import (
	"chattweiler/pkg/vk"
)

type ChatUserJoinEvent struct {
	UserID string
	PeerID int
}

type ChatUserLeavingEvent struct {
	UserID string
	PeerID int
}

type InfoCommand struct {
	PeerID int
}

type ContentRequestCommand struct {
	Type   vk.AttachmentsType
	UserID string
	PeerID int
}
