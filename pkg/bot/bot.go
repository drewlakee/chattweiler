package bot

import (
	"chattweiler/pkg/app/utils"
	"chattweiler/pkg/repository"
	"chattweiler/pkg/repository/model/types"
	"chattweiler/pkg/vk"
	"chattweiler/pkg/vk/events"
	"chattweiler/pkg/vk/messages"
	"chattweiler/pkg/vk/warden/membership"
	"github.com/SevereCloud/vksdk/v2/api"
	vklp "github.com/SevereCloud/vksdk/v2/longpoll-user"
	vklpwrapper "github.com/SevereCloud/vksdk/v2/longpoll-user/v3"
	wrapper "github.com/SevereCloud/vksdk/v2/longpoll-user/v3"
	"github.com/sirupsen/logrus"
	"os"
	"strconv"
	"time"
)

var packageLogFields = logrus.Fields{
	"package": "bot",
}

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
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"func": "NewBot",
			"err":  err,
		}).Fatal("long poll initialization error")
	}

	chatId, err := strconv.ParseInt(os.Getenv("vk.community.chat.id"), 10, 64)
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"func": "NewBot",
			"err":  err,
		}).Fatal("vk.community.chat.id parse failed")
	}

	communityId, err := strconv.ParseInt(os.Getenv("vk.community.id"), 10, 64)
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"func": "NewBot",
			"err":  err,
		}).Fatal("vk.community.id parse failed")
	}

	membershipCheckInterval, err := time.ParseDuration(utils.GetEnvOrDefault("chat.warden.membership.check.interval", "10m"))
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"func": "NewBot",
			"err":  err,
		}).Fatal("chat.warden.membership.check.interval parse failed")
	}

	gracePeriod, err := time.ParseDuration(utils.GetEnvOrDefault("chat.warden.membership.grace.period", "1h"))
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"func": "NewBot",
			"err":  err,
		}).Fatal("chat.warden.membership.grace.period parse failed")
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
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"struct": "Bot",
			"func":   "handleChatUserJoinEvent",
			"err":    err,
		}).Error("message send error")
		return
	}

	logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
		"struct":   "Bot",
		"func":     "handleChatUserJoinEvent",
		"username": user.ScreenName,
	}).Info()
	_, err = bot.vkapi.MessagesSend(messages.BuildMessageUsingPersonalizedPhrase(
		event.PeerID,
		user,
		bot.phrasesRepo.FindAllByType(types.Welcome),
	))
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"struct": "Bot",
			"func":   "handleChatUserJoinEvent",
			"err":    err,
		}).Error("message send error")
	}
}

func (bot *Bot) handleChatUserLeaveEvent(event wrapper.ChatInfoChange) {
	user, err := vk.GetUserInfo(bot.vkapi, event)
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"struct": "Bot",
			"func":   "handleChatUserLeaveEvent",
			"err":    err,
		}).Error("vk api error")
		return
	}

	logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
		"struct":   "Bot",
		"func":     "handleChatUserLeaveEvent",
		"username": user.ScreenName,
	}).Info()
	_, err = bot.vkapi.MessagesSend(messages.BuildMessageUsingPersonalizedPhrase(
		event.PeerID,
		user,
		bot.phrasesRepo.FindAllByType(types.Goodbye),
	))
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"struct": "Bot",
			"func":   "handleChatUserLeaveEvent",
			"err":    err,
		}).Error("message send error")
	}
}

func (bot *Bot) Start() {
	welcomeNewMembers, err := strconv.ParseBool(utils.GetEnvOrDefault("bot.functionality.welcome.new.members", "true"))
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"func": "Start",
			"err":  err,
		}).Fatal("bot.functionality.welcome.new.members parse error")
	}

	goodbyeMembers, err := strconv.ParseBool(utils.GetEnvOrDefault("bot.functionality.goodbye.members", "true"))
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"func": "Start",
			"err":  err,
		}).Fatal("bot.functionality.goodbye.members parse error")
	}

	bot.vklpwrapper.OnChatInfoChange(func(event wrapper.ChatInfoChange) {
		switch events.ResolveChatInfoChangeEventType(event) {
		case vklpwrapper.ChatUserCome:
			if welcomeNewMembers {
				bot.handleChatUserJoinEvent(event)
			}
		case vklpwrapper.ChatUserLeave:
			if goodbyeMembers {
				bot.handleChatUserLeaveEvent(event)
			}
		}
	})

	bot.vklpwrapper.OnNewMessage(func(event wrapper.NewMessage) {
		if event.Text == "bark!" {
			_, err := bot.vkapi.MessagesSend(messages.BuildMessagePhrase(
				event.PeerID,
				bot.phrasesRepo.FindAllByType(types.Info),
			))
			if err != nil {
				logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
					"func": "Start",
					"err":  err,
				}).Error("message send error")
			}
		}
	})

	checkMembership, err := strconv.ParseBool(utils.GetEnvOrDefault("bot.functionality.membership.checking", "true"))
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"func": "Start",
			"err":  err,
		}).Fatal("bot.functionality.membership.checking parse error")
	}

	if checkMembership {
		// run async
		go bot.membershipChecker.LoopCheck()
	}

	logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
		"func": "Start",
	}).Info("Bot is running...")
	err = bot.vklp.Run()
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"func": "Start",
			"err":  err,
		}).Fatal("bot crashed")
	}
}
