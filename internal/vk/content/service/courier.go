package service

import (
	botobject "chattweiler/internal/bot/object"
	"chattweiler/internal/configs"
	"chattweiler/internal/logging"
	"chattweiler/internal/repository"
	"chattweiler/internal/repository/model"
	"chattweiler/internal/utils"
	"chattweiler/internal/vk"
	"chattweiler/internal/vk/content"
	"strconv"
	"time"

	"github.com/SevereCloud/vksdk/v2/api"
	"github.com/SevereCloud/vksdk/v2/object"
)

var logPackage = "service"

type MediaContentCourier struct {
	communityVkApi          *api.VK
	userVkApi               *api.VK
	phrasesRepo             repository.PhraseRepository
	contentCommandRepo      repository.ContentCommandRepository
	listeningChannel        chan *botobject.ContentRequestCommand
	commandCollectors       map[int]content.AttachmentsContentCollector
	garbageCleaningInterval time.Duration
	lastTsGarbageCollected  time.Time
}

func NewMediaContentCourier(
	communityVkApi,
	userVkApi *api.VK,
	phrasesRepo repository.PhraseRepository,
	contentCommandRepo repository.ContentCommandRepository,
	listeningChannel chan *botobject.ContentRequestCommand,
	garbageCleaningInterval time.Duration,
) *MediaContentCourier {
	return &MediaContentCourier{
		communityVkApi:          communityVkApi,
		userVkApi:               userVkApi,
		phrasesRepo:             phrasesRepo,
		contentCommandRepo:      contentCommandRepo,
		listeningChannel:        listeningChannel,
		commandCollectors:       make(map[int]content.AttachmentsContentCollector),
		lastTsGarbageCollected:  time.Now(),
		garbageCleaningInterval: garbageCleaningInterval,
	}
}

func (courier *MediaContentCourier) ReceiveAndDeliver() {
	for {
		select {
		case received := <-courier.listeningChannel:
			user, err := vk.GetUserInfo(courier.communityVkApi, received.Event.UserID)
			if err != nil {
				logging.Log.Error(logPackage, "MediaContentCourier.ReceiveAndDeliver", err, "%s: get user info error", received.Event.UserID)
				continue
			}

			if _, alreadyExists := courier.commandCollectors[received.Command.GetID()]; !alreadyExists {
				courier.createNewCollectorForCommand(received)
			}

			if courier.lastTsGarbageCollected.Add(courier.garbageCleaningInterval).Before(time.Now()) {
				courier.removeGarbageCollectors()
			}

			mediaContent := courier.commandCollectors[received.Command.GetID()].CollectOne()
			if len(mediaContent.Type) == 0 {
				logging.Log.Warn(logPackage, "MediaContentCourier.ReceiveAndDeliver", "collected empty media content ignored")
				courier.askToRetryRequest(received, user)
				continue
			}

			courier.deliverContentResponse(received, user, mediaContent)
		}
	}
}

func (courier *MediaContentCourier) deliverContentResponse(
	request *botobject.ContentRequestCommand,
	user *object.UsersUser,
	mediaContent object.WallWallpostAttachment,
) {
	phrases := courier.phrasesRepo.FindAllByType(model.ContentRequestType)

	var messageToSend api.Params
	if len(phrases) == 0 {
		messageToSend = vk.BuildDirectedMessage(request.Event.PeerID)
	} else {
		messageToSend = vk.BuildMessageUsingPersonalizedPhrase(request.Event.PeerID, user, phrases)
	}

	messageToSend["attachment"] = courier.resolveContentID(mediaContent, request.GetAttachmentsType())
	_, err := courier.communityVkApi.MessagesSend(messageToSend)
	if err != nil {
		logging.Log.Error(logPackage, "MediaContentCourier.deliverContentResponse", err, "message sending error")
	}
}

func (courier *MediaContentCourier) createNewCollectorForCommand(request *botobject.ContentRequestCommand) {
	courier.commandCollectors[request.Command.GetID()] = NewCachedRandomAttachmentsContentCollector(
		courier.userVkApi,
		request.GetAttachmentsType(),
		request.Command.GetID(),
		courier.contentCommandRepo,
		courier.getMaxCachedAttachments(request.GetAttachmentsType()),
		courier.getCacheRefreshThreshold(request.GetAttachmentsType()),
	)
}

func (courier *MediaContentCourier) askToRetryRequest(
	request *botobject.ContentRequestCommand,
	user *object.UsersUser,
) {
	phrases := courier.phrasesRepo.FindAllByType(model.RetryType)
	if len(phrases) == 0 {
		logging.Log.Warn(logPackage, "MediaContentCourier.askToRetryRequest", "there's no ask retry phrases, message won't be sent")
		return
	}

	messageToSend := vk.BuildMessageUsingPersonalizedPhrase(request.Event.PeerID, user, phrases)
	_, err := courier.communityVkApi.MessagesSend(messageToSend)
	if err != nil {
		logging.Log.Error(logPackage, "MediaContentCourier.askToRetryRequest", err, "message sending error")
	}
}

func (courier *MediaContentCourier) getMaxCachedAttachments(mediaType vk.AttachmentsType) int {
	switch mediaType {
	case vk.PhotoType:
		pictureMaxCachedAttachments, err := strconv.ParseInt(utils.GetEnvOrDefault(configs.ContentPictureMaxCachedAttachments), 10, 32)
		if err != nil {
			logging.Log.Panic(logPackage, "MediaContentCourier.getMaxCachedAttachments", err, "%s: parsing of env variable is failed", configs.ContentPictureMaxCachedAttachments.Key)
		}

		return int(pictureMaxCachedAttachments)
	case vk.AudioType:
		audioMaxCachedAttachments, err := strconv.ParseInt(utils.GetEnvOrDefault(configs.ContentAudioMaxCachedAttachments), 10, 32)
		if err != nil {
			logging.Log.Panic(logPackage, "MediaContentCourier.getMaxCachedAttachments", err, "%s: parsing of env variable is failed", configs.ContentAudioMaxCachedAttachments.Key)
		}

		return int(audioMaxCachedAttachments)
	case vk.VideoType:
		videoMaxCachedAttachments, err := strconv.ParseInt(utils.GetEnvOrDefault(configs.ContentVideoMaxCachedAttachments), 10, 32)
		if err != nil {
			logging.Log.Panic(logPackage, "MediaContentCourier.getMaxCachedAttachments", err, "%s: parsing of env variable is failed", configs.ContentVideoMaxCachedAttachments.Key)
		}

		return int(videoMaxCachedAttachments)
	}

	return 0
}

func (courier *MediaContentCourier) getCacheRefreshThreshold(mediaType vk.AttachmentsType) float32 {
	switch mediaType {
	case vk.PhotoType:
		pictureCacheRefreshThreshold, err := strconv.ParseFloat(utils.GetEnvOrDefault(configs.ContentPictureCacheRefreshThreshold), 32)
		if err != nil {
			logging.Log.Panic(logPackage, "MediaContentCourier.getCacheRefreshThreshold", err, "%s: parsing of env variable is failed", configs.ContentPictureCacheRefreshThreshold.Key)
		}

		return float32(pictureCacheRefreshThreshold)
	case vk.AudioType:
		audioCacheRefreshThreshold, err := strconv.ParseFloat(utils.GetEnvOrDefault(configs.ContentAudioCacheRefreshThreshold), 32)
		if err != nil {
			logging.Log.Panic(logPackage, "MediaContentCourier.getCacheRefreshThreshold", err, "%s: parsing of env variable is failed", configs.ContentAudioCacheRefreshThreshold.Key)
		}

		return float32(audioCacheRefreshThreshold)
	case vk.VideoType:
		videoCacheRefreshThreshold, err := strconv.ParseFloat(utils.GetEnvOrDefault(configs.ContentVideoCacheRefreshThreshold), 32)
		if err != nil {
			logging.Log.Panic(logPackage, "MediaContentCourier.getCacheRefreshThreshold", err, "%s: parsing of env variable is failed", configs.ContentAudioCacheRefreshThreshold.Key)
		}

		return float32(videoCacheRefreshThreshold)
	}

	return 0
}

func (courier *MediaContentCourier) resolveContentID(
	mediaContent object.WallWallpostAttachment,
	deliverContentType vk.AttachmentsType,
) string {
	switch deliverContentType {
	case vk.AudioType:
		return mediaContent.Audio.ToAttachment()
	case vk.PhotoType:
		return mediaContent.Photo.ToAttachment()
	case vk.VideoType:
		return mediaContent.Video.ToAttachment()
	}

	return ""
}

func (courier *MediaContentCourier) removeGarbageCollectors() {
	relevantCommands := courier.contentCommandRepo.FindAll()
	relevantCommandsMap := make(map[int]bool, len(relevantCommands))
	for _, command := range relevantCommands {
		relevantCommandsMap[command.GetID()] = true
	}

	for commandID := range courier.commandCollectors {
		if _, exist := relevantCommandsMap[commandID]; !exist {
			delete(courier.commandCollectors, commandID)
		}
	}
}
