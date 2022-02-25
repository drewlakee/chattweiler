package bot

import (
	"chattweiler/pkg/repository"
	"chattweiler/pkg/repository/model/types"
	"chattweiler/pkg/vk"
	"chattweiler/pkg/vk/events"
	"chattweiler/pkg/vk/messages"
	"fmt"
	"github.com/SevereCloud/vksdk/v2/api"
	vklp "github.com/SevereCloud/vksdk/v2/longpoll-user"
	vklpwrapper "github.com/SevereCloud/vksdk/v2/longpoll-user/v3"
	wrapper "github.com/SevereCloud/vksdk/v2/longpoll-user/v3"
)

type Bot struct {
	vkapi                  *api.VK
	vklp                   *vklp.LongPoll
	vklpwrapper            *wrapper.Wrapper
	phraseRepo             repository.PhraseRepository
	membershipWarningsRepo repository.MembershipWarningRepository
}

func NewBot(vkToken string, phraseRepo repository.PhraseRepository, membershipWarningsRepo repository.MembershipWarningRepository) *Bot {
	vkapi := api.NewVK(vkToken)

	lp, err := vklp.NewLongPoll(vkapi, 0)
	if err != nil {
		panic(err)
	}

	wrappedlp := vklpwrapper.NewWrapper(lp)

	return &Bot{
		vkapi:                  vkapi,
		vklp:                   lp,
		vklpwrapper:            wrappedlp,
		phraseRepo:             phraseRepo,
		membershipWarningsRepo: membershipWarningsRepo,
	}
}

func (bot *Bot) handleChatUserJoinEvent(event wrapper.ChatInfoChange) {
	user, err := vk.GetUserInfo(bot.vkapi, event)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("User", user.ScreenName, "has joined to the chat")
	_, err = bot.vkapi.MessagesSend(messages.BuildMessageUsingPersonalizedPhrase(
		event.PeerID,
		user,
		bot.phraseRepo.FindAllByType(types.Welcome),
	))
	if err != nil {
		fmt.Println(err)
		return
	}
}

func (bot *Bot) handleChatUserLeaveEvent(event wrapper.ChatInfoChange) {
	user, err := vk.GetUserInfo(bot.vkapi, event)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("User", user.ScreenName, "has leaved the chat")
	_, err = bot.vkapi.MessagesSend(messages.BuildMessageUsingPersonalizedPhrase(
		event.PeerID,
		user,
		bot.phraseRepo.FindAllByType(types.Goodbye),
	))
	if err != nil {
		fmt.Println(err)
		return
	}
}

func (bot *Bot) Start() error {
	bot.vklpwrapper.OnChatInfoChange(func(event wrapper.ChatInfoChange) {
		switch events.ResolveChatInfoChangeEventType(event) {
		case vklpwrapper.ChatUserCome:
			bot.handleChatUserJoinEvent(event)
		case vklpwrapper.ChatUserLeave:
			bot.handleChatUserLeaveEvent(event)
		}
	})

	fmt.Println("Bot is running...")
	err := bot.vklp.Run()
	if err != nil {
		return err
	}

	return nil
}
