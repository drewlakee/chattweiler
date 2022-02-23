package main

import (
	"fmt"
	vkapi "github.com/SevereCloud/vksdk/v2/api"
	vklp "github.com/SevereCloud/vksdk/v2/longpoll-user"
	vklpwrapper "github.com/SevereCloud/vksdk/v2/longpoll-user/v3"
	"os"
)

func main() {
	vk := vkapi.NewVK(os.Getenv("vk_community_bot_token"))

	poll, err := vklp.NewLongPoll(vk, 0)
	if err != nil {
		panic(err)
	}

	wrappedLongPoll := vklpwrapper.NewWrapper(poll)
	wrappedLongPoll.OnChatInfoChange(func(event vklpwrapper.ChatInfoChange) {
		if event.TypeID-1 == vklpwrapper.ChatUserCome {
			response, _ := vk.UsersGet(vkapi.Params{
				"user_ids": event.Info,
				"fields":   "screen_name",
			})
			fmt.Println("User", response[0].ScreenName, "has joined to the chat")
			_, _ = vk.MessagesSend(vkapi.Params{
				"peer_id":   event.PeerID,
				"random_id": 0,
				"message":   "Hey there, @" + response[0].ScreenName + "!",
			})
		}

		if event.TypeID-1 == vklpwrapper.ChatUserLeave {
			response, _ := vk.UsersGet(vkapi.Params{
				"user_ids": event.Info,
				"fields":   "screen_name",
			})
			fmt.Println("User", response[0].ScreenName, "has leaved the chat")
			_, _ = vk.MessagesSend(vkapi.Params{
				"peer_id":   event.PeerID,
				"random_id": 0,
				"message":   "See you later @" + response[0].ScreenName + "!",
			})
		}
	})

	err = poll.Run()
	if err != nil {
		panic(err)
	}
}
