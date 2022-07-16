package service

import (
	botobject "chattweiler/pkg/bot/object"
	"chattweiler/pkg/configs"
	"chattweiler/pkg/repository"
	"chattweiler/pkg/repository/model"
	"chattweiler/pkg/utils"
	"chattweiler/pkg/vk"
	"chattweiler/pkg/vk/content"
	"strconv"
	"time"

	"github.com/SevereCloud/vksdk/v2/api"
	"github.com/SevereCloud/vksdk/v2/object"
	"github.com/sirupsen/logrus"
)

var packageLogFields = logrus.Fields{
	"package": "service",
}

type MediaContentCourier struct {
	communityVkApi          *api.VK
	userVkApi               *api.VK
	phrasesRepo             repository.PhraseRepository
	contentCommandRepo      repository.ContentCommandRepository
	listeningChannel        chan *botobject.ContentRequestCommand
	commandCollectors       map[string]content.AttachmentsContentCollector
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
		commandCollectors:       make(map[string]content.AttachmentsContentCollector),
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
				logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
					"struct": "MediaContentCourier",
					"func":   "ReceiveAndDeliver",
					"err":    err,
					"user":   received.Event.UserID,
				}).Error("get user info error")
				continue
			}

			if _, alreadyExists := courier.commandCollectors[received.Command.Name]; !alreadyExists {
				courier.createNewCollectorForCommand(received)
			}

			if courier.lastTsGarbageCollected.Add(courier.garbageCleaningInterval).Before(time.Now()) {
				courier.removeGarbageCollectors()
			}

			mediaContent := courier.commandCollectors[received.Command.Name].CollectOne()
			if len(mediaContent.Type) == 0 {
				logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
					"func":   "ReceiveAndDeliver",
					"struct": "MediaContentCourier",
				}).Warn("collected empty media content ignored")
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
	messageToSend := courier.getResponseMessage(request, user)
	messageToSend["attachment"] = courier.resolveContentID(mediaContent, request.GetAttachmentsType())
	_, err := courier.communityVkApi.MessagesSend(messageToSend)
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"func":   "deliverContentResponse",
			"struct": "MediaContentCourier",
			"err":    err,
		}).Error("message send error")
	}
}

func (courier *MediaContentCourier) createNewCollectorForCommand(request *botobject.ContentRequestCommand) {
	courier.commandCollectors[request.Command.Name] = NewCachedRandomAttachmentsContentCollector(
		courier.userVkApi,
		request.GetAttachmentsType(),
		request.Command.Name,
		courier.contentCommandRepo,
		courier.getMaxCachedAttachments(request.GetAttachmentsType()),
		courier.getCacheRefreshThreshold(request.GetAttachmentsType()),
	)
}

func (courier *MediaContentCourier) askToRetryRequest(
	request *botobject.ContentRequestCommand,
	user *object.UsersUser,
) {
	messageToSend := vk.BuildMessageUsingPersonalizedPhrase(
		request.Event.PeerID,
		user,
		model.RetryType,
		courier.phrasesRepo.FindAllByType(model.RetryType),
	)

	if messageToSend["message"] != nil {
		_, err := courier.communityVkApi.MessagesSend(messageToSend)
		if err != nil {
			logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
				"func":   "deliverContentResponse",
				"struct": "MediaContentCourier",
				"err":    err,
			}).Error("message send error")
		}
	}
}

func (courier *MediaContentCourier) getResponseMessage(
	request *botobject.ContentRequestCommand,
	user *object.UsersUser,
) api.Params {
	return vk.BuildMessageUsingPersonalizedPhrase(
		request.Event.PeerID,
		user,
		model.ContentRequestType,
		courier.phrasesRepo.FindAllByType(model.ContentRequestType),
	)
}

func (courier *MediaContentCourier) getMaxCachedAttachments(mediaType vk.AttachmentsType) int {
	switch mediaType {
	case vk.PhotoType:
		pictureMaxCachedAttachments, err := strconv.ParseInt(utils.GetEnvOrDefault(configs.ContentPictureMaxCachedAttachments), 10, 32)
		if err != nil {
			logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
				"struct": "MediaContentCourier",
				"func":   "getMaxCachedAttachments",
				"type":   vk.PhotoType,
				"err":    err,
				"key":    configs.ContentPictureMaxCachedAttachments.Key,
			}).Fatal("parsing of env variable is failed")
		}

		return int(pictureMaxCachedAttachments)
	case vk.AudioType:
		audioMaxCachedAttachments, err := strconv.ParseInt(utils.GetEnvOrDefault(configs.ContentAudioMaxCachedAttachments), 10, 32)
		if err != nil {
			logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
				"struct": "MediaContentCourier",
				"func":   "getMaxCachedAttachments",
				"type":   vk.AudioType,
				"err":    err,
				"key":    configs.ContentAudioMaxCachedAttachments.Key,
			}).Fatal("parsing of env variable is failed")
		}

		return int(audioMaxCachedAttachments)
	case vk.VideoType:
		videoMaxCachedAttachments, err := strconv.ParseInt(utils.GetEnvOrDefault(configs.ContentVideoMaxCachedAttachments), 10, 32)
		if err != nil {
			logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
				"struct": "MediaContentCourier",
				"func":   "getMaxCachedAttachments",
				"type":   vk.VideoType,
				"err":    err,
				"key":    configs.ContentAudioMaxCachedAttachments.Key,
			}).Fatal("parsing of env variable is failed")
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
			logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
				"struct": "MediaContentCourier",
				"func":   "getCacheRefreshThreshold",
				"type":   vk.PhotoType,
				"err":    err,
				"key":    configs.ContentPictureCacheRefreshThreshold.Key,
			}).Fatal("parsing of env variable is failed")
		}

		return float32(pictureCacheRefreshThreshold)
	case vk.AudioType:
		audioCacheRefreshThreshold, err := strconv.ParseFloat(utils.GetEnvOrDefault(configs.ContentAudioCacheRefreshThreshold), 32)
		if err != nil {
			logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
				"struct": "MediaContentCourier",
				"func":   "getCacheRefreshThreshold",
				"type":   vk.AudioType,
				"err":    err,
				"key":    configs.ContentAudioCacheRefreshThreshold.Key,
			}).Fatal("parsing of env variable is failed")
		}

		return float32(audioCacheRefreshThreshold)
	case vk.VideoType:
		videoCacheRefreshThreshold, err := strconv.ParseFloat(utils.GetEnvOrDefault(configs.ContentVideoCacheRefreshThreshold), 32)
		if err != nil {
			logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
				"struct": "MediaContentCourier",
				"func":   "getCacheRefreshThreshold",
				"type":   vk.VideoType,
				"err":    err,
				"key":    configs.ContentAudioCacheRefreshThreshold.Key,
			}).Fatal("parsing of env variable is failed")
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
	relevantCommandsMap := make(map[string]bool, len(relevantCommands))
	for _, command := range relevantCommands {
		relevantCommandsMap[command.Name] = true
	}

	for commandName, _ := range courier.commandCollectors {
		if _, exist := relevantCommandsMap[commandName]; !exist {
			delete(courier.commandCollectors, commandName)
		}
	}
}
