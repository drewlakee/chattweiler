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
	vkapi       *api.VK
	vklp        *vklp.LongPoll
	vklpwrapper *wrapper.Wrapper

	phrasesRepo        repository.PhraseRepository
	contentCommandRepo repository.ContentCommandRepository

	membershipChecker *vk.Checker

	contentCommandInputChannel chan *object.ContentRequestCommand
	contentCourier             *service.MediaContentCourier
}

func NewLongPoolingBot(
	phrasesRepo repository.PhraseRepository,
	membershipWarningsRepo repository.MembershipWarningRepository,
	contentCommandRepo repository.ContentCommandRepository,
) *LongPoolingBot {
	vkBotToken := utils.MustGetEnv(configs.VkCommunityBotToken)
	communityVkApi := api.NewVK(vkBotToken)

	mode := vklp.ReceiveAttachments + vklp.ExtendedEvents
	lp, err := vklp.NewLongPoll(communityVkApi, mode)
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"func": "NewLongPoolingBot",
			"err":  err,
		}).Fatal("long poll initialization error")
	}

	chatId, err := strconv.ParseInt(utils.MustGetEnv(configs.VkCommunityChatID), 10, 64)
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"func": "NewLongPoolingBot",
			"err":  err,
			"key":  configs.VkCommunityChatID.Key,
		}).Fatal("parsing of env variable is failed")
	}

	communityId, err := strconv.ParseInt(utils.MustGetEnv(configs.VkCommunityID), 10, 64)
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"func": "NewLongPoolingBot",
			"err":  err,
			"key":  configs.VkCommunityID.Key,
		}).Fatal("parsing of env variable is failed")
	}

	membershipCheckInterval, err := time.ParseDuration(utils.GetEnvOrDefault(configs.ChatWarderMembershipCheckInterval))
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"func": "NewLongPoolingBot",
			"err":  err,
			"key":  configs.ChatWarderMembershipCheckInterval.Key,
		}).Fatal("parsing of env variable is failed")
	}

	gracePeriod, err := time.ParseDuration(utils.GetEnvOrDefault(configs.ChatWardenMembershipGracePeriod))
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"func": "NewLongPoolingBot",
			"err":  err,
			"key":  configs.ChatWardenMembershipGracePeriod.Key,
		}).Fatal("parsing of env variable is failed")
	}

	vkUserApi := api.NewVK(utils.GetEnvOrDefault(configs.VkAdminUserToken))

	requestsQueueSize, err := strconv.ParseInt(utils.GetEnvOrDefault(configs.ContentRequestsQueueSize), 10, 32)
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"func": "NewLongPoolingBot",
			"err":  err,
			"key":  configs.ContentPictureQueueSize.Key,
		}).Fatal("parsing of env variable is failed")
	}

	contentRequestsInputChannel := make(chan *object.ContentRequestCommand, requestsQueueSize)

	garbageCollectorsCleaningInterval, err := time.ParseDuration(utils.GetEnvOrDefault(configs.ContentGarbageCollectorsCleaningInterval))
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"func": "NewBot",
			"err":  err,
			"key":  configs.ContentPictureQueueSize.Key,
		}).Fatal("parsing of env variable is failed")
	}

	return &LongPoolingBot{
		vkapi:                      communityVkApi,
		vklp:                       lp,
		vklpwrapper:                vklpwrapper.NewWrapper(lp),
		phrasesRepo:                phrasesRepo,
		contentCommandRepo:         contentCommandRepo,
		membershipChecker:          vk.NewChecker(chatId, communityId, membershipCheckInterval, gracePeriod, communityVkApi, phrasesRepo, membershipWarningsRepo),
		contentCourier:             service.NewMediaContentCourier(communityVkApi, vkUserApi, phrasesRepo, contentCommandRepo, contentRequestsInputChannel, garbageCollectorsCleaningInterval),
		contentCommandInputChannel: contentRequestsInputChannel,
	}
}

func (bot *LongPoolingBot) Serve() {
	welcomeNewMembers, err := strconv.ParseBool(utils.GetEnvOrDefault(configs.BotFunctionalityWelcomeNewMembers))
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"func": "Serve",
			"err":  err,
			"key":  configs.BotFunctionalityWelcomeNewMembers.Key,
		}).Fatal("parsing of env variable is failed")
	}

	goodbyeMembers, err := strconv.ParseBool(utils.GetEnvOrDefault(configs.BotFunctionalityGoodbyeMembers))
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"func": "Serve",
			"err":  err,
			"key":  configs.BotFunctionalityGoodbyeMembers.Key,
		}).Fatal("parsing of env variable is failed")
	}

	bot.vklpwrapper.OnChatInfoChange(func(event wrapper.ChatInfoChange) {
		switch resolveChatInfoChangeEventType(event) {
		case vklpwrapper.ChatUserCome:
			if welcomeNewMembers {
				bot.handleChatUserJoinEvent(mapper.NewChatEventFromFromChatInfoChange(event))
			}
		case vklpwrapper.ChatUserLeave:
			if goodbyeMembers {
				bot.handleChatUserLeavingEvent(mapper.NewChatEventFromFromChatInfoChange(event))
			}
		}
	})

	contentRequests, err := strconv.ParseBool(utils.GetEnvOrDefault(configs.BotFunctionalityContentCommands))
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"func": "Serve",
			"err":  err,
			"key":  configs.BotFunctionalityContentCommands.Key,
		}).Fatal("parsing of env variable is failed")
	}

	if contentRequests {
		// run async
		go bot.contentCourier.ReceiveAndDeliver()
	}

	infoCommand := utils.GetEnvOrDefault(configs.BotCommandOverrideInfo)
	bot.vklpwrapper.OnNewMessage(func(event wrapper.NewMessage) {
		switch event.Text {
		case infoCommand:
			bot.handleInfoCommand(mapper.NewChatEventFromNewMessage(event))
		default:
			if contentRequests {
				if contentCommand := bot.contentCommandRepo.FindByCommand(event.Text); contentCommand != nil {
					bot.handleContentRequestCommand(mapper.NewContentCommandRequest(contentCommand, event))
				}
			}
		}
	})

	bot.startMembershipCheckingAsync()

	logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
		"func": "Serve",
	}).Info("Bot is running...")
	err = bot.vklp.Run()
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"func": "Serve",
			"err":  err,
		}).Fatal("bot is crashed")
	}
}

func (bot *LongPoolingBot) handleChatUserJoinEvent(event *object.ChatEvent) {
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

	messageToSend := vk.BuildMessageUsingPersonalizedPhrase(
		event.PeerID,
		user,
		model.WelcomeType,
		bot.phrasesRepo.FindAllByType(model.WelcomeType),
	)

	if _, messageContainsPhrase := messageToSend["message"]; !messageContainsPhrase {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"struct": "Bot",
			"func":   "handleChatUserJoinEvent",
		}).Warn("message doesn't have any phrase for a join event")
		return
	}

	_, err = bot.vkapi.MessagesSend(messageToSend)
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"struct": "Bot",
			"func":   "handleChatUserJoinEvent",
			"err":    err,
		}).Error("message sending error")
	}
}

func (bot *LongPoolingBot) handleChatUserLeavingEvent(event *object.ChatEvent) {
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

	messageToSend := vk.BuildMessageUsingPersonalizedPhrase(
		event.PeerID,
		user,
		model.GoodbyeType,
		bot.phrasesRepo.FindAllByType(model.GoodbyeType),
	)

	if _, messageContainsPhrase := messageToSend["message"]; !messageContainsPhrase {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"struct": "Bot",
			"func":   "handleChatUserLeavingEvent",
		}).Warn("message doesn't have any phrase for a leaving event")
		return
	}

	_, err = bot.vkapi.MessagesSend(messageToSend)
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"struct": "Bot",
			"func":   "handleChatUserLeavingEvent",
			"err":    err,
		}).Error("message sending error")
	}
}

func (bot *LongPoolingBot) handleInfoCommand(event *object.ChatEvent) {
	messageToSend := vk.BuildMessagePhrase(
		event.PeerID,
		bot.phrasesRepo.FindAllByType(model.InfoType),
	)

	if messageToSend["message"] != nil {
		_, err := bot.vkapi.MessagesSend(messageToSend)
		if err != nil {
			logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
				"func": "handleInfoCommand",
				"err":  err,
			}).Error("message sending error")
		}
	}
}

func (bot *LongPoolingBot) handleContentRequestCommand(request *object.ContentRequestCommand) {
	bot.contentCommandInputChannel <- request
}

func (bot *LongPoolingBot) startMembershipCheckingAsync() {
	checkMembership, err := strconv.ParseBool(utils.GetEnvOrDefault(configs.BotFunctionalityMembershipChecking))
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"func": "checkMembership",
			"err":  err,
			"key":  configs.BotFunctionalityMembershipChecking.Key,
		}).Fatal("parsing of env variable is failed")
	}

	if checkMembership {
		go bot.membershipChecker.LoopCheck()
	}
}

func resolveChatInfoChangeEventType(event wrapper.ChatInfoChange) wrapper.TypeID {
	return event.TypeID - 1
}
