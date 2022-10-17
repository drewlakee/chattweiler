package bot

import (
	"chattweiler/pkg/bot/object"
	"chattweiler/pkg/bot/object/mapper"
	"chattweiler/pkg/configs"
	"chattweiler/pkg/logging"
	"chattweiler/pkg/repository"
	"chattweiler/pkg/repository/model"
	"chattweiler/pkg/utils"
	"chattweiler/pkg/vk"
	"chattweiler/pkg/vk/content/service"
	"strconv"
	"strings"
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
		logging.Log.Panic(logPackage, "NewLongPoolingBot", err, "long-poll initialization error")
	}

	chatId, err := strconv.ParseInt(utils.MustGetEnv(configs.VkCommunityChatID), 10, 64)
	if err != nil {
		logging.Log.Panic(logPackage, "NewLongPoolingBot", err, "%s: parsing of env variable is failed", configs.VkCommunityChatID.Key)
	}

	communityId, err := strconv.ParseInt(utils.MustGetEnv(configs.VkCommunityID), 10, 64)
	if err != nil {
		logging.Log.Panic(logPackage, "NewLongPoolingBot", err, "%s: parsing of env variable is failed", configs.VkCommunityID.Key)
	}

	membershipCheckInterval, err := time.ParseDuration(utils.GetEnvOrDefault(configs.ChatWarderMembershipCheckInterval))
	if err != nil {
		logging.Log.Panic(logPackage, "NewLongPoolingBot", err, "%s: parsing of env variable is failed", configs.ChatWarderMembershipCheckInterval.Key)
	}

	gracePeriod, err := time.ParseDuration(utils.GetEnvOrDefault(configs.ChatWardenMembershipGracePeriod))
	if err != nil {
		logging.Log.Panic(logPackage, "NewLongPoolingBot", err, "%s: parsing of env variable is failed", configs.ChatWardenMembershipGracePeriod.Key)
	}

	vkUserApi := api.NewVK(utils.GetEnvOrDefault(configs.VkAdminUserToken))

	requestsQueueSize, err := strconv.ParseInt(utils.GetEnvOrDefault(configs.ContentRequestsQueueSize), 10, 32)
	if err != nil {
		logging.Log.Panic(logPackage, "NewLongPoolingBot", err, "%s: parsing of env variable is failed", configs.ContentRequestsQueueSize.Key)
	}

	contentRequestsInputChannel := make(chan *object.ContentRequestCommand, requestsQueueSize)

	garbageCollectorsCleaningInterval, err := time.ParseDuration(utils.GetEnvOrDefault(configs.ContentGarbageCollectorsCleaningInterval))
	if err != nil {
		logging.Log.Panic(logPackage, "NewLongPoolingBot", err, "%s: parsing of env variable is failed", configs.ContentGarbageCollectorsCleaningInterval.Key)
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
	welcomeNewMembersFeatureEnabled, err := strconv.ParseBool(utils.GetEnvOrDefault(configs.BotFunctionalityWelcomeNewMembers))
	if err != nil {
		logging.Log.Panic(logPackage, "LongPoolingBot.Serve", err, "%s: parsing of env variable is failed", configs.BotFunctionalityWelcomeNewMembers.Key)
	}

	goodbyeMembersFeatureEnabled, err := strconv.ParseBool(utils.GetEnvOrDefault(configs.BotFunctionalityGoodbyeMembers))
	if err != nil {
		logging.Log.Panic(logPackage, "LongPoolingBot.Serve", err, "%s: parsing of env variable is failed", configs.BotFunctionalityGoodbyeMembers.Key)
	}

	bot.vklpwrapper.OnChatInfoChange(func(event wrapper.ChatInfoChange) {
		switch resolveChatInfoChangeEventType(event) {
		case vklpwrapper.ChatUserCome:
			if welcomeNewMembersFeatureEnabled {
				bot.handleChatUserJoinEvent(mapper.NewChatEventFromFromChatInfoChange(event))
			}
		case vklpwrapper.ChatUserLeave:
			if goodbyeMembersFeatureEnabled {
				bot.handleChatUserLeavingEvent(mapper.NewChatEventFromFromChatInfoChange(event))
			}
		}
	})

	contentRequestsFeatureEnabled, err := strconv.ParseBool(utils.GetEnvOrDefault(configs.BotFunctionalityContentCommands))
	if err != nil {
		logging.Log.Panic(logPackage, "LongPoolingBot.Serve", err, "%s: parsing of env variable is failed", configs.BotFunctionalityContentCommands.Key)
	}

	if contentRequestsFeatureEnabled {
		// run async
		go bot.contentCourier.ReceiveAndDeliver()
	}

	infoCommand := utils.GetEnvOrDefault(configs.BotCommandOverrideInfo)
	bot.vklpwrapper.OnNewMessage(func(event wrapper.NewMessage) {
		if strings.EqualFold(event.Text, infoCommand) {
			bot.handleInfoCommand(mapper.NewChatEventFromNewMessage(event))
		}

		if !contentRequestsFeatureEnabled {
			return
		}

		if contentCommand := bot.contentCommandRepo.FindByCommandAlias(event.Text); contentCommand != nil {
			bot.handleContentRequestCommand(mapper.NewContentCommandRequest(contentCommand, event))
		}
	})

	// run async
	bot.startMembershipCheckingAsync()

	logging.Log.Info(logPackage, "LongPoolingBot.Serve", "Bot is running...")
	err = bot.vklp.Run()
	if err != nil {
		logging.Log.Panic(logPackage, "LongPoolingBot.Serve", err, "bot is crashed")
	}
}

func (bot *LongPoolingBot) handleChatUserJoinEvent(event *object.ChatEvent) {
	user, err := vk.GetUserInfo(bot.vkapi, event.UserID)
	if err != nil {
		logging.Log.Error(logPackage, "LongPoolingBot.handleChatUserJoinEvent", err, "message sending error")
		return
	}

	logging.Log.Info(logPackage, "LongPoolingBot.handleChatUserJoinEvent", "'%s' user is joined", user.ScreenName)
	messageToSend := vk.BuildMessageUsingPersonalizedPhrase(
		event.PeerID,
		user,
		model.WelcomeType,
		bot.phrasesRepo.FindAllByType(model.WelcomeType),
	)

	if _, messageContainsPhrase := messageToSend["message"]; !messageContainsPhrase {
		logging.Log.Warn(logPackage, "LongPoolingBot.handleChatUserJoinEvent", "message doesn't have any phrase to send for '%s'", user.ScreenName)
		return
	}

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
	messageToSend := vk.BuildMessageUsingPersonalizedPhrase(
		event.PeerID,
		user,
		model.GoodbyeType,
		bot.phrasesRepo.FindAllByType(model.GoodbyeType),
	)

	if _, messageContainsPhrase := messageToSend["message"]; !messageContainsPhrase {
		logging.Log.Warn(logPackage, "LongPoolingBot.handleChatUserLeavingEvent", "message doesn't have any phrase to send for '%s'", user.ScreenName)
		return
	}

	_, err = bot.vkapi.MessagesSend(messageToSend)
	if err != nil {
		logging.Log.Error(logPackage, "LongPoolingBot.handleChatUserLeavingEvent", err, "message sending error")
	}
}

func (bot *LongPoolingBot) handleInfoCommand(event *object.ChatEvent) {
	messageToSend := vk.BuildMessagePhrase(
		event.PeerID,
		bot.phrasesRepo.FindAllByType(model.InfoType),
	)

	if messageToSend["message"] != nil && messageToSend["message"] != "" {
		_, err := bot.vkapi.MessagesSend(messageToSend)
		if err != nil {
			logging.Log.Error(logPackage, "LongPoolingBot.handleInfoCommand", err, "message sending error")
		}
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

func resolveChatInfoChangeEventType(event wrapper.ChatInfoChange) wrapper.TypeID {
	// some framework specific workaround - messed event order 😵‍
	return event.TypeID - 1
}
