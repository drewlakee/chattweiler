package events

import wrapper "github.com/SevereCloud/vksdk/v2/longpoll-user/v3"

func ResolveChatInfoChangeEventType(event wrapper.ChatInfoChange) wrapper.TypeID {
	return event.TypeID - 1
}
