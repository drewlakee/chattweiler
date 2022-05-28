package bot

import (
	"chattweiler/pkg/app/configs/static"
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
	"strconv"
	"time"
)

var packageLogFields = logrus.Fields{
	"package": "bot",
}

type BotCommand string

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
	phrasesRepo repository.PhraseRepository,
	membershipWarningsRepo repository.MembershipWarningRepository,
	contentSourceRepo repository.ContentSourceRepository,
) *Bot {
	vkBotToken := utils.MustGetEnv(static.VkCommunityBotToken)
	communityVkApi := api.NewVK(vkBotToken)

	mode := vklp.ReceiveAttachments + vklp.ExtendedEvents
	lp, err := vklp.NewLongPoll(communityVkApi, mode)
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"func": "NewBot",
			"err":  err,
		}).Fatal("long poll initialization error")
	}

	chatId, err := strconv.ParseInt(utils.MustGetEnv(static.VkCommunityChatID), 10, 64)
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"func": "NewBot",
			"err":  err,
			"key":  static.VkCommunityChatID.Key,
		}).Fatal("parsing of env variable is failed")
	}

	communityId, err := strconv.ParseInt(utils.MustGetEnv(static.VkCommunityID), 10, 64)
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"func": "NewBot",
			"err":  err,
			"key":  static.VkCommunityID.Key,
		}).Fatal("parsing of env variable is failed")
	}

	membershipCheckInterval, err := time.ParseDuration(utils.GetEnvOrDefault(static.ChatWarderMembershipCheckInterval))
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"func": "NewBot",
			"err":  err,
			"key":  static.ChatWarderMembershipCheckInterval.Key,
		}).Fatal("parsing of env variable is failed")
	}

	gracePeriod, err := time.ParseDuration(utils.GetEnvOrDefault(static.ChatWardenMembershipGracePeriod))
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"func": "NewBot",
			"err":  err,
			"key":  static.ChatWardenMembershipGracePeriod.Key,
		}).Fatal("parsing of env variable is failed")
	}

	vkUserApi := api.NewVK(utils.GetEnvOrDefault(static.VkAdminUserToken))
	audioMaxCachedAttachments, err := strconv.ParseInt(utils.GetEnvOrDefault(static.ContentAudioMaxCachedAttachments), 10, 32)
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"func": "NewBot",
			"err":  err,
			"key":  static.ContentAudioMaxCachedAttachments.Key,
		}).Fatal("parsing of env variable is failed")
	}

	audioCacheRefreshThreshold, err := strconv.ParseFloat(utils.GetEnvOrDefault(static.ContentAudioCacheRefreshThreshold), 32)
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"func": "NewBot",
			"err":  err,
			"key":  static.ContentAudioCacheRefreshThreshold.Key,
		}).Fatal("parsing of env variable is failed")
	}

	audioCollector := random.NewCachedRandomAttachmentsContentCollector(vkUserApi, content.Audio, contentSourceRepo, int(audioMaxCachedAttachments), float32(audioCacheRefreshThreshold))

	audioQueueSize, err := strconv.ParseInt(utils.GetEnvOrDefault(static.ContentAudioQueueSize), 10, 32)
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"func": "NewBot",
			"err":  err,
			"key":  static.ContentAudioQueueSize.Key,
		}).Fatal("parsing of env variable is failed")
	}

	pictureQueueSize, err := strconv.ParseInt(utils.GetEnvOrDefault(static.ContentPictureQueueSize), 10, 32)
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"func": "NewBot",
			"err":  err,
			"key":  static.ContentPictureQueueSize.Key,
		}).Fatal("content.picture.queue.size parsing is failed")
	}

	pictureMaxCachedAttachments, err := strconv.ParseInt(utils.GetEnvOrDefault(static.ContentPictureMaxCachedAttachments), 10, 32)
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"func": "NewBot",
			"err":  err,
			"key":  static.ContentPictureMaxCachedAttachments.Key,
		}).Fatal("parsing of env variable is failed")
	}

	pictureCacheRefreshThreshold, err := strconv.ParseFloat(utils.GetEnvOrDefault(static.ContentPictureCacheRefreshThreshold), 32)
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"func": "NewBot",
			"err":  err,
			"key":  static.ContentPictureCacheRefreshThreshold.Key,
		}).Fatal("parsing of env variable is failed")
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
	welcomeNewMembers, err := strconv.ParseBool(utils.GetEnvOrDefault(static.BotFunctionalityWelcomeNewMembers))
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"func": "Start",
			"err":  err,
			"key":  static.BotFunctionalityWelcomeNewMembers.Key,
		}).Fatal("parsing of env variable is failed")
	}

	goodbyeMembers, err := strconv.ParseBool(utils.GetEnvOrDefault(static.BotFunctionalityGoodbyeMembers))
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"func": "Start",
			"err":  err,
			"key":  static.BotFunctionalityGoodbyeMembers.Key,
		}).Fatal("parsing of env variable is failed")
	}

	bot.vklpwrapper.OnChatInfoChange(func(event wrapper.ChatInfoChange) {
		switch events.ResolveChatInfoChangeEventType(event) {
		case vklpwrapper.ChatUserCome:
			if welcomeNewMembers {
				bot.handleChatUserJoinEvent(event)
			}
		case vklpwrapper.ChatUserLeave:
			if goodbyeMembers {
				bot.handleChatUserLeavingEvent(event)
			}
		}
	})

	audioRequests, err := strconv.ParseBool(utils.GetEnvOrDefault(static.BotFunctionalityAudioRequests))
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"func": "Start",
			"err":  err,
			"key":  static.BotFunctionalityGoodbyeMembers.Key,
		}).Fatal("parsing of env variable is failed")
	}

	if audioRequests {
		// run async
		go bot.audioContentCourier.ReceiveAndDeliver(types.AudioRequest, content.Audio, bot.audioContentCourierChannel)
	}

	pictureRequests, err := strconv.ParseBool(utils.GetEnvOrDefault(static.BotFunctionalityPictureRequests))
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"func": "Start",
			"err":  err,
			"key":  static.BotFunctionalityPictureRequests.Key,
		}).Fatal("parsing of env variable is failed")
	}

	if pictureRequests {
		// run async
		go bot.pictureContentCourier.ReceiveAndDeliver(types.PictureRequest, content.Photo, bot.pictureContentCourierChannel)
	}

	infoCommand := utils.GetEnvOrDefault(static.BotCommandOverrideInfo)
	audioRequestCommand := utils.GetEnvOrDefault(static.BotCommandOverrideAudioRequest)
	pictureRequestCommand := utils.GetEnvOrDefault(static.BotCommandOverridePictureRequest)

	logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
		"func": "Start",
		"call": infoCommand,
	}).Info("info command registered")

	if audioRequests {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"func": "Start",
			"call": audioRequestCommand,
		}).Info("audio request command registered")
	}

	if pictureRequests {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"func": "Start",
			"call": pictureRequestCommand,
		}).Info("picture request command registered")
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

	checkMembership, err := strconv.ParseBool(utils.GetEnvOrDefault(static.BotFunctionalityMembershipChecking))
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"func": "Start",
			"err":  err,
			"key":  static.BotFunctionalityMembershipChecking.Key,
		}).Fatal("parsing of env variable is failed")
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
		}).Fatal("bot is crashed")
	}
}

func (bot *Bot) handleChatUserJoinEvent(event wrapper.ChatInfoChange) {
	user, err := vk.GetUserInfo(bot.vkapi, strconv.Itoa(event.Info))
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"struct": "Bot",
			"func":   "handleChatUserJoinEvent",
			"err":    err,
		}).Error("message sending error")
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
		types.Welcome,
		bot.phrasesRepo.FindAllByType(types.Welcome),
	)

	if _, messageContainsPhrase := params["message"]; !messageContainsPhrase {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"struct": "Bot",
			"func":   "handleChatUserJoinEvent",
		}).Warn("message doesn't have any phrase for a join event")
		return
	}

	_, err = bot.vkapi.MessagesSend(params)
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"struct": "Bot",
			"func":   "handleChatUserJoinEvent",
			"err":    err,
		}).Error("message sending error")
	}
}

func (bot *Bot) handleChatUserLeavingEvent(event wrapper.ChatInfoChange) {
	user, err := vk.GetUserInfo(bot.vkapi, strconv.Itoa(event.Info))
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"struct": "Bot",
			"func":   "handleChatUserLeavingEvent",
			"err":    err,
		}).Error("vk api error")
		return
	}

	logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
		"struct":   "Bot",
		"func":     "handleChatUserLeavingEvent",
		"username": user.ScreenName,
	}).Info()

	params := messages.BuildMessageUsingPersonalizedPhrase(
		event.PeerID,
		user,
		types.Goodbye,
		bot.phrasesRepo.FindAllByType(types.Goodbye),
	)

	if _, messageContainsPhrase := params["message"]; !messageContainsPhrase {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"struct": "Bot",
			"func":   "handleChatUserLeavingEvent",
		}).Warn("message doesn't have any phrase for a leaving event")
		return
	}

	_, err = bot.vkapi.MessagesSend(params)
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"struct": "Bot",
			"func":   "handleChatUserLeavingEvent",
			"err":    err,
		}).Error("message sending error")
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
		}).Error("message sending error")
	}
}

func (bot *Bot) handleAudioRequestCommand(audioRequest wrapper.NewMessage) {
	bot.audioContentCourierChannel <- audioRequest
}

func (bot *Bot) handlePictureRequestCommand(pictureRequest wrapper.NewMessage) {
	bot.pictureContentCourierChannel <- pictureRequest
}
