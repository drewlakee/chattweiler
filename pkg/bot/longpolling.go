package bot

import (
	"chattweiler/pkg/bot/object"
	"chattweiler/pkg/bot/object/mapper"
	"chattweiler/pkg/configs"
	"chattweiler/pkg/repository"
	"chattweiler/pkg/repository/model"
	"chattweiler/pkg/utils"
	"chattweiler/pkg/vk"
	"chattweiler/pkg/vk/content/service"
	"strconv"
	"time"

	"github.com/SevereCloud/vksdk/v2/api"

	vklp "github.com/SevereCloud/vksdk/v2/longpoll-user"
	vklpwrapper "github.com/SevereCloud/vksdk/v2/longpoll-user/v3"
	wrapper "github.com/SevereCloud/vksdk/v2/longpoll-user/v3"
	"github.com/sirupsen/logrus"
)

type LongPoolingBot struct {
	vkapi                        *api.VK
	vklp                         *vklp.LongPoll
	vklpwrapper                  *wrapper.Wrapper
	phrasesRepo                  repository.PhraseRepository
	membershipChecker            *vk.Checker
	audioContentCourier          *service.MediaContentCourier
	audioContentCourierChannel   chan *object.ContentRequestCommand
	pictureContentCourier        *service.MediaContentCourier
	pictureContentCourierChannel chan *object.ContentRequestCommand
}

func NewLongPoolingBot(
	phrasesRepo repository.PhraseRepository,
	membershipWarningsRepo repository.MembershipWarningRepository,
	contentSourceRepo repository.ContentSourceRepository,
) *LongPoolingBot {
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

	audioCollector := service.NewCachedRandomAttachmentsContentCollector(vkUserApi, vk.AudioType, contentSourceRepo, int(audioMaxCachedAttachments), float32(audioCacheRefreshThreshold))

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
		}).Fatal("service.picture.queue.size parsing is failed")
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

	pictureCollector := service.NewCachedRandomAttachmentsContentCollector(vkUserApi, vk.PhotoType, contentSourceRepo, int(pictureMaxCachedAttachments), float32(pictureCacheRefreshThreshold))

	return &LongPoolingBot{
		vkapi:                        communityVkApi,
		vklp:                         lp,
		vklpwrapper:                  vklpwrapper.NewWrapper(lp),
		phrasesRepo:                  phrasesRepo,
		membershipChecker:            vk.NewChecker(chatId, communityId, membershipCheckInterval, gracePeriod, communityVkApi, phrasesRepo, membershipWarningsRepo),
		audioContentCourier:          service.NewMediaContentCourier(communityVkApi, vkUserApi, audioCollector, phrasesRepo),
		audioContentCourierChannel:   make(chan *object.ContentRequestCommand, audioQueueSize),
		pictureContentCourier:        service.NewMediaContentCourier(communityVkApi, vkUserApi, pictureCollector, phrasesRepo),
		pictureContentCourierChannel: make(chan *object.ContentRequestCommand, pictureQueueSize),
	}
}

func (bot *LongPoolingBot) Serve() {
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
				bot.handleChatUserJoinEvent(mapper.NewChatUserJoinEventFromChatInfoChange(event))
			}
		case vklpwrapper.ChatUserLeave:
			if goodbyeMembers {
				bot.handleChatUserLeavingEvent(mapper.NewChatUserLeavingEventFromChatInfoChange(event))
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
		go bot.audioContentCourier.ReceiveAndDeliver(model.AudioRequestType, bot.audioContentCourierChannel)
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
		go bot.audioContentCourier.ReceiveAndDeliver(model.PictureRequestType, bot.pictureContentCourierChannel)
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
			bot.handleInfoCommand(mapper.NewInfoCommandFromNewMessage(event))
		case audioRequestCommand:
			if audioRequests {
				bot.handleAudioRequestCommand(mapper.NewContentRequestCommandFromNewMessage(vk.AudioType, event))
			}
		case pictureRequestCommand:
			if pictureRequests {
				bot.handlePictureRequestCommand(mapper.NewContentRequestCommandFromNewMessage(vk.PhotoType, event))
			}
		}
	})

	bot.checkMembership()

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

func (bot *LongPoolingBot) handleChatUserJoinEvent(event *object.ChatUserJoinEvent) {
	user, err := vk.GetUserInfo(bot.vkapi, event.UserID)
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

func (bot *LongPoolingBot) handleChatUserLeavingEvent(event *object.ChatUserLeavingEvent) {
	user, err := vk.GetUserInfo(bot.vkapi, event.UserID)
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

func (bot *LongPoolingBot) handleInfoCommand(event *object.InfoCommand) {
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

func (bot *LongPoolingBot) handleAudioRequestCommand(request *object.ContentRequestCommand) {
	bot.audioContentCourierChannel <- request
}

func (bot *LongPoolingBot) handlePictureRequestCommand(request *object.ContentRequestCommand) {
	bot.pictureContentCourierChannel <- request
}

func (bot *LongPoolingBot) checkMembership() {
	checkMembership, err := strconv.ParseBool(utils.GetEnvOrDefault(configs.BotFunctionalityMembershipChecking))
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"func": "checkMembership",
			"err":  err,
			"key":  configs.BotFunctionalityMembershipChecking.Key,
		}).Fatal("parsing of env variable is failed")
	}

	if checkMembership {
		// run async
		go bot.membershipChecker.LoopCheck()
	}
}
