package bot

import (
	"chattweiler/pkg/configs"
	"chattweiler/pkg/repository"
	"chattweiler/pkg/repository/model"
	"chattweiler/pkg/utils"
	"chattweiler/pkg/vk"
	"chattweiler/pkg/vk/content"
	"strconv"
	"time"

	"github.com/SevereCloud/vksdk/v2/api"

	vklp "github.com/SevereCloud/vksdk/v2/longpoll-user"
	vklpwrapper "github.com/SevereCloud/vksdk/v2/longpoll-user/v3"
	wrapper "github.com/SevereCloud/vksdk/v2/longpoll-user/v3"
	"github.com/sirupsen/logrus"
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
	membershipChecker            *vk.Checker
	audioContentCourier          *content.MediaContentCourier
	audioContentCourierChannel   chan wrapper.NewMessage
	pictureContentCourier        *content.MediaContentCourier
	pictureContentCourierChannel chan wrapper.NewMessage
}

func NewBot(
	phrasesRepo repository.PhraseRepository,
	membershipWarningsRepo repository.MembershipWarningRepository,
	contentSourceRepo repository.ContentSourceRepository,
) *Bot {
	vkBotToken := utils.MustGetEnv(configs.VkCommunityBotToken)
	communityVkApi := api.NewVK(vkBotToken)

	mode := vklp.ReceiveAttachments + vklp.ExtendedEvents
	lp, err := vklp.NewLongPoll(communityVkApi, mode)
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"func": "NewBot",
			"err":  err,
		}).Fatal("long poll initialization error")
	}

	chatId, err := strconv.ParseInt(utils.MustGetEnv(configs.VkCommunityChatID), 10, 64)
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"func": "NewBot",
			"err":  err,
			"key":  configs.VkCommunityChatID.Key,
		}).Fatal("parsing of env variable is failed")
	}

	communityId, err := strconv.ParseInt(utils.MustGetEnv(configs.VkCommunityID), 10, 64)
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"func": "NewBot",
			"err":  err,
			"key":  configs.VkCommunityID.Key,
		}).Fatal("parsing of env variable is failed")
	}

	membershipCheckInterval, err := time.ParseDuration(utils.GetEnvOrDefault(configs.ChatWarderMembershipCheckInterval))
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"func": "NewBot",
			"err":  err,
			"key":  configs.ChatWarderMembershipCheckInterval.Key,
		}).Fatal("parsing of env variable is failed")
	}

	gracePeriod, err := time.ParseDuration(utils.GetEnvOrDefault(configs.ChatWardenMembershipGracePeriod))
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"func": "NewBot",
			"err":  err,
			"key":  configs.ChatWardenMembershipGracePeriod.Key,
		}).Fatal("parsing of env variable is failed")
	}

	vkUserApi := api.NewVK(utils.GetEnvOrDefault(configs.VkAdminUserToken))
	audioMaxCachedAttachments, err := strconv.ParseInt(utils.GetEnvOrDefault(configs.ContentAudioMaxCachedAttachments), 10, 32)
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"func": "NewBot",
			"err":  err,
			"key":  configs.ContentAudioMaxCachedAttachments.Key,
		}).Fatal("parsing of env variable is failed")
	}

	audioCacheRefreshThreshold, err := strconv.ParseFloat(utils.GetEnvOrDefault(configs.ContentAudioCacheRefreshThreshold), 32)
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"func": "NewBot",
			"err":  err,
			"key":  configs.ContentAudioCacheRefreshThreshold.Key,
		}).Fatal("parsing of env variable is failed")
	}

	audioCollector := content.NewCachedRandomAttachmentsContentCollector(vkUserApi, content.AudioType, contentSourceRepo, int(audioMaxCachedAttachments), float32(audioCacheRefreshThreshold))

	audioQueueSize, err := strconv.ParseInt(utils.GetEnvOrDefault(configs.ContentAudioQueueSize), 10, 32)
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"func": "NewBot",
			"err":  err,
			"key":  configs.ContentAudioQueueSize.Key,
		}).Fatal("parsing of env variable is failed")
	}

	pictureQueueSize, err := strconv.ParseInt(utils.GetEnvOrDefault(configs.ContentPictureQueueSize), 10, 32)
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"func": "NewBot",
			"err":  err,
			"key":  configs.ContentPictureQueueSize.Key,
		}).Fatal("content.picture.queue.size parsing is failed")
	}

	pictureMaxCachedAttachments, err := strconv.ParseInt(utils.GetEnvOrDefault(configs.ContentPictureMaxCachedAttachments), 10, 32)
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"func": "NewBot",
			"err":  err,
			"key":  configs.ContentPictureMaxCachedAttachments.Key,
		}).Fatal("parsing of env variable is failed")
	}

	pictureCacheRefreshThreshold, err := strconv.ParseFloat(utils.GetEnvOrDefault(configs.ContentPictureCacheRefreshThreshold), 32)
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"func": "NewBot",
			"err":  err,
			"key":  configs.ContentPictureCacheRefreshThreshold.Key,
		}).Fatal("parsing of env variable is failed")
	}

	pictureCollector := content.NewCachedRandomAttachmentsContentCollector(vkUserApi, content.PhotoType, contentSourceRepo, int(pictureMaxCachedAttachments), float32(pictureCacheRefreshThreshold))

	return &Bot{
		vkapi:                        communityVkApi,
		vklp:                         lp,
		vklpwrapper:                  vklpwrapper.NewWrapper(lp),
		phrasesRepo:                  phrasesRepo,
		membershipChecker:            vk.NewChecker(chatId, communityId, membershipCheckInterval, gracePeriod, communityVkApi, phrasesRepo, membershipWarningsRepo),
		audioContentCourier:          content.NewMediaContentCourier(communityVkApi, vkUserApi, audioCollector, phrasesRepo),
		audioContentCourierChannel:   make(chan wrapper.NewMessage, audioQueueSize),
		pictureContentCourier:        content.NewMediaContentCourier(communityVkApi, vkUserApi, pictureCollector, phrasesRepo),
		pictureContentCourierChannel: make(chan wrapper.NewMessage, pictureQueueSize),
	}
}

func (bot *Bot) Start() {
	welcomeNewMembers, err := strconv.ParseBool(utils.GetEnvOrDefault(configs.BotFunctionalityWelcomeNewMembers))
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"func": "Start",
			"err":  err,
			"key":  configs.BotFunctionalityWelcomeNewMembers.Key,
		}).Fatal("parsing of env variable is failed")
	}

	goodbyeMembers, err := strconv.ParseBool(utils.GetEnvOrDefault(configs.BotFunctionalityGoodbyeMembers))
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"func": "Start",
			"err":  err,
			"key":  configs.BotFunctionalityGoodbyeMembers.Key,
		}).Fatal("parsing of env variable is failed")
	}

	bot.vklpwrapper.OnChatInfoChange(func(event wrapper.ChatInfoChange) {
		switch vk.ResolveChatInfoChangeEventType(event) {
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

	audioRequests, err := strconv.ParseBool(utils.GetEnvOrDefault(configs.BotFunctionalityAudioRequests))
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"func": "Start",
			"err":  err,
			"key":  configs.BotFunctionalityGoodbyeMembers.Key,
		}).Fatal("parsing of env variable is failed")
	}

	if audioRequests {
		// run async
		go bot.audioContentCourier.ReceiveAndDeliver(model.AudioRequestType, content.AudioType, bot.audioContentCourierChannel)
	}

	pictureRequests, err := strconv.ParseBool(utils.GetEnvOrDefault(configs.BotFunctionalityPictureRequests))
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"func": "Start",
			"err":  err,
			"key":  configs.BotFunctionalityPictureRequests.Key,
		}).Fatal("parsing of env variable is failed")
	}

	if pictureRequests {
		// run async
		go bot.audioContentCourier.ReceiveAndDeliver(model.PictureRequestType, content.PhotoType, bot.pictureContentCourierChannel)
	}

	infoCommand := utils.GetEnvOrDefault(configs.BotCommandOverrideInfo)
	audioRequestCommand := utils.GetEnvOrDefault(configs.BotCommandOverrideAudioRequest)
	pictureRequestCommand := utils.GetEnvOrDefault(configs.BotCommandOverridePictureRequest)

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

	checkMembership, err := strconv.ParseBool(utils.GetEnvOrDefault(configs.BotFunctionalityMembershipChecking))
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"func": "Start",
			"err":  err,
			"key":  configs.BotFunctionalityMembershipChecking.Key,
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

	params := vk.BuildMessageUsingPersonalizedPhrase(
		event.PeerID,
		user,
		model.WelcomeType,
		bot.phrasesRepo.FindAllByType(model.WelcomeType),
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

	params := vk.BuildMessageUsingPersonalizedPhrase(
		event.PeerID,
		user,
		model.GoodbyeType,
		bot.phrasesRepo.FindAllByType(model.GoodbyeType),
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
	_, err := bot.vkapi.MessagesSend(vk.BuildMessagePhrase(
		event.PeerID,
		bot.phrasesRepo.FindAllByType(model.InfoType),
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
