package bot

import (
	"chattweiler/pkg/repository"
	"chattweiler/pkg/repository/model/types"
	"chattweiler/pkg/vk"
	"chattweiler/pkg/vk/events"
	"chattweiler/pkg/vk/messages"
	"chattweiler/pkg/vk/warden/membership"
	"errors"
	"fmt"
	"github.com/SevereCloud/vksdk/v2/api"
	vklp "github.com/SevereCloud/vksdk/v2/longpoll-user"
	vklpwrapper "github.com/SevereCloud/vksdk/v2/longpoll-user/v3"
	wrapper "github.com/SevereCloud/vksdk/v2/longpoll-user/v3"
	"os"
	"strconv"
	"time"
)

type Bot struct {
	vkapi             *api.VK
	vklp              *vklp.LongPoll
	vklpwrapper       *wrapper.Wrapper
	phrasesRepo       repository.PhraseRepository
	membershipChecker *membership.Checker
}

func NewBot(vkToken string, phrasesRepo repository.PhraseRepository, membershipWarningsRepo repository.MembershipWarningRepository) *Bot {
	vkapi := api.NewVK(vkToken)

	lp, err := vklp.NewLongPoll(vkapi, 0)
	if err != nil {
		panic(err)
	}

	chatId, err := strconv.ParseInt(os.Getenv("vk.community.chat.id"), 10, 64)
	if err != nil {
		fmt.Println(err)
		panic(errors.New("Membership checker initialization: vk.community.chat.id parse failed"))
	}

	communityId, err := strconv.ParseInt(os.Getenv("vk.community.id"), 10, 64)
	if err != nil {
		fmt.Println(err)
		panic(errors.New("Membership checker initialization: vk.community.id parse failed"))
	}

	rawMembershipCheckInterval := os.Getenv("chat.warden.membership.check.interval")
	membershipCheckInterval, err := time.ParseDuration(rawMembershipCheckInterval)
	if err != nil {
		fmt.Println(err)
		panic(errors.New("Membership checker initialization: chat.warden.membership.check.interval parse failed"))
	}

	gracePeriod, err := time.ParseDuration(os.Getenv("chat.warden.membership.grace.period"))
	if err != nil {
		fmt.Println(err)
		panic(errors.New("Membership checker initialization: chat.warden.membership.grace.period parse failed"))
	}

	return &Bot{
		vkapi:             vkapi,
		vklp:              lp,
		vklpwrapper:       vklpwrapper.NewWrapper(lp),
		phrasesRepo:       phrasesRepo,
		membershipChecker: membership.NewChecker(chatId, communityId, membershipCheckInterval, gracePeriod, vkapi, phrasesRepo, membershipWarningsRepo),
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
		bot.phrasesRepo.FindAllByType(types.Welcome),
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
		bot.phrasesRepo.FindAllByType(types.Goodbye),
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

	bot.vklpwrapper.OnNewMessage(func(event wrapper.NewMessage) {
		if event.Text == "bark!" {
			_, err := bot.vkapi.MessagesSend(messages.BuildMessagePhrase(
				event.PeerID,
				bot.phrasesRepo.FindAllByType(types.Info),
			))
			if err != nil {
				fmt.Println(err)
				return
			}
		}
	})

	// run async
	go bot.membershipChecker.LoopCheck()

	fmt.Println("Bot is running...")
	err := bot.vklp.Run()
	if err != nil {
		return err
	}

	return nil
}
