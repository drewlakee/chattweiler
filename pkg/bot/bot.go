package bot

import (
	"chattweiler/pkg/app/utils"
	"chattweiler/pkg/repository"
	"chattweiler/pkg/repository/model/types"
	"chattweiler/pkg/vk"
	"chattweiler/pkg/vk/content"
	"chattweiler/pkg/vk/content/random"
	"chattweiler/pkg/vk/courier"
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

type BotCommand string

const (
	Info           BotCommand = "bark!"
	AudioRequest   BotCommand = "sing song!"
	PictureRequest BotCommand = "gimme pic!"
)

type Bot struct {
	vkapi                        *api.VK
	vklp                         *vklp.LongPoll
	vklpwrapper                  *wrapper.Wrapper
	phrasesRepo                  repository.PhraseRepository
	membershipChecker            *membership.Checker
	audioContentCourier          *courier.MediaContentCourier
	audioContentCourierChannel   chan wrapper.NewMessage
	pictureContentCourier        *courier.MediaContentCourier
	pictureContentCourierChannel chan wrapper.NewMessage
}

func NewBot(
	vkToken string,
	phrasesRepo repository.PhraseRepository,
	membershipWarningsRepo repository.MembershipWarningRepository,
	contentSourceRepo repository.ContentSourceRepository,
) *Bot {
	communityVkApi := api.NewVK(vkToken)

	mode := vklp.ReceiveAttachments + vklp.ExtendedEvents
	lp, err := vklp.NewLongPoll(communityVkApi, mode)
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

	vkUserApi := api.NewVK(utils.GetEnvOrDefault("vk.admin.user.token", ""))
	audioMaxCachedAttachments, err := strconv.ParseInt(utils.GetEnvOrDefault("content.audio.max.cached.attachments", "100"), 10, 32)
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"func": "NewBot",
			"err":  err,
		}).Fatal("content.audio.max.cached.attachments parse failed")
	}

	audioCacheRefreshThreshold, err := strconv.ParseFloat(utils.GetEnvOrDefault("content.audio.cache.refresh.threshold", "0.2"), 32)
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"func": "NewBot",
			"err":  err,
		}).Fatal("content.audio.cache.refresh.threshold parse failed")
	}

	audioCollector := random.NewCachedRandomAttachmentsContentCollector(vkUserApi, content.Audio, contentSourceRepo, int(audioMaxCachedAttachments), float32(audioCacheRefreshThreshold))

	audioQueueSize, err := strconv.ParseInt(utils.GetEnvOrDefault("content.audio.queue.size", "100"), 10, 32)
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"func": "NewBot",
			"err":  err,
		}).Fatal("content.audio.queue.size parse failed")
	}

	pictureQueueSize, err := strconv.ParseInt(utils.GetEnvOrDefault("content.picture.queue.size", "100"), 10, 32)
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"func": "NewBot",
			"err":  err,
		}).Fatal("content.picture.queue.size parse failed")
	}

	pictureMaxCachedAttachments, err := strconv.ParseInt(utils.GetEnvOrDefault("content.picture.max.cached.attachments", "100"), 10, 32)
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"func": "NewBot",
			"err":  err,
		}).Fatal("content.picture.max.cached.attachments parse failed")
	}

	pictureCacheRefreshThreshold, err := strconv.ParseFloat(utils.GetEnvOrDefault("content.picture.cache.refresh.threshold", "0.2"), 32)
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"func": "NewBot",
			"err":  err,
		}).Fatal("content.picture.cache.refresh.threshold parse failed")
	}

	pictureCollector := random.NewCachedRandomAttachmentsContentCollector(vkUserApi, content.Photo, contentSourceRepo, int(pictureMaxCachedAttachments), float32(pictureCacheRefreshThreshold))

	return &Bot{
		vkapi:                        communityVkApi,
		vklp:                         lp,
		vklpwrapper:                  vklpwrapper.NewWrapper(lp),
		phrasesRepo:                  phrasesRepo,
		membershipChecker:            membership.NewChecker(chatId, communityId, membershipCheckInterval, gracePeriod, communityVkApi, phrasesRepo, membershipWarningsRepo),
		audioContentCourier:          courier.NewMediaContentCourier(communityVkApi, vkUserApi, audioCollector, phrasesRepo),
		audioContentCourierChannel:   make(chan wrapper.NewMessage, audioQueueSize),
		pictureContentCourier:        courier.NewMediaContentCourier(communityVkApi, vkUserApi, pictureCollector, phrasesRepo),
		pictureContentCourierChannel: make(chan wrapper.NewMessage, pictureQueueSize),
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

	audioRequests, err := strconv.ParseBool(utils.GetEnvOrDefault("bot.functionality.audio.requests", "false"))
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"func": "Start",
			"err":  err,
		}).Fatal("bot.functionality.audio.requests parse error")
	}

	if audioRequests {
		// run async
		go bot.audioContentCourier.ReceiveAndDeliver(types.AudioRequest, content.Audio, bot.audioContentCourierChannel)
	}

	pictureRequests, err := strconv.ParseBool(utils.GetEnvOrDefault("bot.functionality.picture.requests", "false"))
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"func": "Start",
			"err":  err,
		}).Fatal("bot.functionality.picture.requests parse error")
	}

	if pictureRequests {
		// run async
		go bot.pictureContentCourier.ReceiveAndDeliver(types.PictureRequest, content.Photo, bot.pictureContentCourierChannel)
	}

	infoCommand := utils.GetEnvOrDefault("bot.command.override.info", string(Info))
	audioRequestCommand := utils.GetEnvOrDefault("bot.command.override.audio.request", string(AudioRequest))
	pictureRequestCommand := utils.GetEnvOrDefault("bot.command.override.picture.request", string(PictureRequest))

	logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
		"func":      "Start",
		"call_name": infoCommand,
	}).Info("info command")

	if audioRequests {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"func":      "Start",
			"call_name": audioRequestCommand,
		}).Info("audio request command")
	}

	if pictureRequests {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"func":      "Start",
			"call_name": pictureRequestCommand,
		}).Info("picture request command")
	}

	bot.vklpwrapper.OnNewMessage(func(event wrapper.NewMessage) {
		switch event.Text {
		case infoCommand:
			bot.handleInfoCommand(event)
		case audioRequestCommand:
			if audioRequests {
				bot.handleAudioRequestCommand(event)
			}
		case pictureRequestCommand:
			if pictureRequests {
				bot.handlePictureRequestCommand(event)
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

func (bot *Bot) handleChatUserJoinEvent(event wrapper.ChatInfoChange) {
	user, err := vk.GetUserInfo(bot.vkapi, strconv.Itoa(event.Info))
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

	params := messages.BuildMessageUsingPersonalizedPhrase(
		event.PeerID,
		user,
		bot.phrasesRepo.FindAllByType(types.Welcome),
	)

	if len(params) == 0 {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"struct": "Bot",
			"func":   "handleChatUserJoinEvent",
		}).Warn("empty api params ignored")
		return
	}

	_, err = bot.vkapi.MessagesSend(params)
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"struct": "Bot",
			"func":   "handleChatUserJoinEvent",
			"err":    err,
		}).Error("message send error")
	}
}

func (bot *Bot) handleChatUserLeaveEvent(event wrapper.ChatInfoChange) {
	user, err := vk.GetUserInfo(bot.vkapi, strconv.Itoa(event.Info))
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

	params := messages.BuildMessageUsingPersonalizedPhrase(
		event.PeerID,
		user,
		bot.phrasesRepo.FindAllByType(types.Goodbye),
	)

	if len(params) == 0 {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"struct": "Bot",
			"func":   "handleChatUserLeaveEvent",
		}).Warn("empty api params ignored")
		return
	}

	_, err = bot.vkapi.MessagesSend(params)
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"struct": "Bot",
			"func":   "handleChatUserLeaveEvent",
			"err":    err,
		}).Error("message send error")
	}
}

func (bot *Bot) handleInfoCommand(event wrapper.NewMessage) {
	_, err := bot.vkapi.MessagesSend(messages.BuildMessagePhrase(
		event.PeerID,
		bot.phrasesRepo.FindAllByType(types.Info),
	))
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"func": "handleInfoCommand",
			"err":  err,
		}).Error("message send error")
	}
}

func (bot *Bot) handleAudioRequestCommand(audioRequest wrapper.NewMessage) {
	bot.audioContentCourierChannel <- audioRequest
}

func (bot *Bot) handlePictureRequestCommand(pictureRequest wrapper.NewMessage) {
	bot.pictureContentCourierChannel <- pictureRequest
}
