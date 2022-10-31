package bot

import (
	"chattweiler/internal/bot/object"
	"chattweiler/internal/bot/object/mapper"
	"chattweiler/internal/configs"
	"chattweiler/internal/logging"
	"chattweiler/internal/repository"
	"chattweiler/internal/repository/model"
	"chattweiler/internal/utils"
	"chattweiler/internal/vk"
	"chattweiler/internal/vk/content/service"
	"strconv"
	"time"

	"github.com/SevereCloud/vksdk/v2/api"

	vklp "github.com/SevereCloud/vksdk/v2/longpoll-user"
	vklpwrapper "github.com/SevereCloud/vksdk/v2/longpoll-user/v3"
	wrapper "github.com/SevereCloud/vksdk/v2/longpoll-user/v3"
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

	welcomeNewMembersFeatureEnabled bool
	goodbyeMembersFeatureEnabled    bool
	contentRequestsFeatureEnabled   bool
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
	panicIfError(err, "NewLongPoolingBot", "long-poll initialization error")

	chatId, err := strconv.ParseInt(utils.MustGetEnv(configs.VkCommunityChatID), 10, 64)
	panicIfError(err, "NewLongPoolingBot", "%s: parsing of env variable is failed", configs.VkCommunityChatID.Key)

	communityId, err := strconv.ParseInt(utils.MustGetEnv(configs.VkCommunityID), 10, 64)
	panicIfError(err, "NewLongPoolingBot", "%s: parsing of env variable is failed", configs.VkCommunityID.Key)

	membershipCheckInterval, err := time.ParseDuration(utils.GetEnvOrDefault(configs.ChatWarderMembershipCheckInterval))
	panicIfError(err, "NewLongPoolingBot", "%s: parsing of env variable is failed", configs.ChatWarderMembershipCheckInterval.Key)

	gracePeriod, err := time.ParseDuration(utils.GetEnvOrDefault(configs.ChatWardenMembershipGracePeriod))
	panicIfError(err, "NewLongPoolingBot", "%s: parsing of env variable is failed", configs.ChatWardenMembershipGracePeriod.Key)

	requestsQueueSize, err := strconv.ParseInt(utils.GetEnvOrDefault(configs.ContentRequestsQueueSize), 10, 32)
	panicIfError(err, "NewLongPoolingBot", "%s: parsing of env variable is failed", configs.ContentRequestsQueueSize.Key)

	welcomeNewMembersFeatureEnabled, err := strconv.ParseBool(utils.GetEnvOrDefault(configs.BotFunctionalityWelcomeNewMembers))
	panicIfError(err, "NewLongPoolingBot", "%s: parsing of env variable is failed", configs.BotFunctionalityWelcomeNewMembers.Key)

	goodbyeMembersFeatureEnabled, err := strconv.ParseBool(utils.GetEnvOrDefault(configs.BotFunctionalityGoodbyeMembers))
	panicIfError(err, "NewLongPoolingBot", "%s: parsing of env variable is failed", configs.BotFunctionalityGoodbyeMembers.Key)

	contentRequestsFeatureEnabled, err := strconv.ParseBool(utils.GetEnvOrDefault(configs.BotFunctionalityContentCommands))
	panicIfError(err, "NewLongPoolingBot", "%s: parsing of env variable is failed", configs.BotFunctionalityContentCommands.Key)

	contentRequestsInputChannel := make(chan *object.ContentRequestCommand, requestsQueueSize)

	garbageCollectorsCleaningInterval, err := time.ParseDuration(utils.GetEnvOrDefault(configs.ContentGarbageCollectorsCleaningInterval))
	panicIfError(err, "%s: parsing of env variable is failed", configs.ContentGarbageCollectorsCleaningInterval.Key)

	vkUserApi := api.NewVK(utils.GetEnvOrDefault(configs.VkAdminUserToken))
	vklWrapper := vklpwrapper.NewWrapper(lp)
	membershipChecker := vk.NewChecker(chatId, communityId, membershipCheckInterval, gracePeriod, communityVkApi, phrasesRepo, membershipWarningsRepo)
	contentCourier := service.NewMediaContentCourier(communityVkApi, vkUserApi, phrasesRepo, contentCommandRepo, contentRequestsInputChannel, garbageCollectorsCleaningInterval)

	return &LongPoolingBot{
		vkapi:                           communityVkApi,
		vklp:                            lp,
		vklpwrapper:                     vklWrapper,
		phrasesRepo:                     phrasesRepo,
		contentCommandRepo:              contentCommandRepo,
		membershipChecker:               membershipChecker,
		contentCourier:                  contentCourier,
		contentCommandInputChannel:      contentRequestsInputChannel,
		welcomeNewMembersFeatureEnabled: welcomeNewMembersFeatureEnabled,
		goodbyeMembersFeatureEnabled:    goodbyeMembersFeatureEnabled,
		contentRequestsFeatureEnabled:   contentRequestsFeatureEnabled,
	}
}

func (bot *LongPoolingBot) Serve() {
	bot.vklpwrapper.OnChatInfoChange(func(event wrapper.ChatInfoChange) {
		switch resolveChatInfoChangeEventType(event) {
		case vklpwrapper.ChatUserCome:
			if bot.welcomeNewMembersFeatureEnabled {
				bot.handleChatUserJoinEvent(mapper.NewChatEventFromFromChatInfoChange(event))
			}
		case vklpwrapper.ChatUserLeave:
			if bot.goodbyeMembersFeatureEnabled {
				bot.handleChatUserLeavingEvent(mapper.NewChatEventFromFromChatInfoChange(event))
			}
		}
	})

	if bot.contentRequestsFeatureEnabled {
		// run async
		go bot.contentCourier.ReceiveAndDeliver()
	}

	bot.vklpwrapper.OnNewMessage(func(event wrapper.NewMessage) {
		command := bot.contentCommandRepo.FindByCommandAlias(event.Text)
		if command == nil {
			return
		}

		switch command.Type {
		case model.InfoCommand:
			bot.handleInfoCommand(mapper.NewChatEventFromNewMessage(event))
		case model.ContentCommand:
			if bot.contentRequestsFeatureEnabled {
				bot.handleContentRequestCommand(mapper.NewContentCommandRequest(command, event))
			}
		}
	})

	// run async
	bot.startMembershipCheckingAsync()

	logging.Log.Info(logPackage, "LongPoolingBot.Serve", "Bot is running...")
	err := bot.vklp.Run()
	panicIfError(err, "LongPoolingBot.Serve", "bot is crashed")
}

func (bot *LongPoolingBot) handleChatUserJoinEvent(event *object.ChatEvent) {
	user, err := vk.GetUserInfo(bot.vkapi, event.UserID)
	if err != nil {
		logging.Log.Error(logPackage, "LongPoolingBot.handleChatUserJoinEvent", err, "message sending error")
		return
	}

	logging.Log.Info(logPackage, "LongPoolingBot.handleChatUserJoinEvent", "'%s' user is joined", user.ScreenName)
	phrases := bot.phrasesRepo.FindAllByType(model.WelcomeType)
	if len(phrases) == 0 {
		logging.Log.Warn(logPackage, "LongPoolingBot.handleChatUserJoinEvent", "there's no welcome phrases, message won't be sent")
		return
	}

	messageToSend := vk.BuildMessageUsingPersonalizedPhrase(event.PeerID, user, phrases)
	_, err = bot.vkapi.MessagesSend(messageToSend)
	if err != nil {
		logging.Log.Error(logPackage, "LongPoolingBot.handleChatUserJoinEvent", err, "message sending error")
	}
}

func (bot *LongPoolingBot) handleChatUserLeavingEvent(event *object.ChatEvent) {
	user, err := vk.GetUserInfo(bot.vkapi, event.UserID)
	if err != nil {
		logging.Log.Error(logPackage, "LongPoolingBot.handleChatUserLeavingEvent", err, "vk api error")
		return
	}

	logging.Log.Info(logPackage, "LongPoolingBot.handleChatUserLeavingEvent", "'%s' user is gone", user.ScreenName)
	phrases := bot.phrasesRepo.FindAllByType(model.GoodbyeType)
	if len(phrases) == 0 {
		logging.Log.Warn(logPackage, "LongPoolingBot.handleChatUserJoinEvent", "there's no goodbye phrases, message won't be sent")
		return
	}

	messageToSend := vk.BuildMessageUsingPersonalizedPhrase(event.PeerID, user, phrases)
	_, err = bot.vkapi.MessagesSend(messageToSend)
	if err != nil {
		logging.Log.Error(logPackage, "LongPoolingBot.handleChatUserLeavingEvent", err, "message sending error. Sent params: %v", messageToSend)
	}
}

func (bot *LongPoolingBot) handleInfoCommand(event *object.ChatEvent) {
	phrases := bot.phrasesRepo.FindAllByType(model.InfoType)
	if len(phrases) == 0 {
		logging.Log.Warn(logPackage, "LongPoolingBot.handleInfoCommand", "there's no info phrases, message won't be sent")
		return
	}

	messageToSend := vk.BuildMessageWithRandomPhrase(event.PeerID, phrases)
	_, err := bot.vkapi.MessagesSend(messageToSend)
	if err != nil {
		logging.Log.Error(logPackage, "LongPoolingBot.handleInfoCommand", err, "message sending error. Sent params: %v", messageToSend)
	}
}

func (bot *LongPoolingBot) handleContentRequestCommand(request *object.ContentRequestCommand) {
	bot.contentCommandInputChannel <- request
}

func (bot *LongPoolingBot) startMembershipCheckingAsync() {
	checkMembershipFeatureEnabled, err := strconv.ParseBool(utils.GetEnvOrDefault(configs.BotFunctionalityMembershipChecking))
	if err != nil {
		logging.Log.Panic(logPackage, "LongPoolingBot.startMembershipCheckingAsync", err, "%s: parsing of env variable is failed", configs.BotFunctionalityMembershipChecking.Key)
	}

	if checkMembershipFeatureEnabled {
		go bot.membershipChecker.LoopCheck()
	}
}

func panicIfError(err error, funcName, messageFormat string, args ...interface{}) {
	if err != nil {
		logging.Log.Panic(logPackage, funcName, err, messageFormat, args)
	}
}

func resolveChatInfoChangeEventType(event wrapper.ChatInfoChange) wrapper.TypeID {
	// some framework specific workaround - messed event order 😵‍
	return event.TypeID - 1
}
