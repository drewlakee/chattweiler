// Package provides object sturctures which bot uses
// to operate
package object

import (
	"chattweiler/pkg/vk"
)

type ChatEvent struct {
	UserID string
	PeerID int
}

type ContentRequestCommand struct {
	Type         vk.AttachmentsType
	RequestEvent *ChatEvent
}
