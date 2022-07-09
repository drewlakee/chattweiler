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

	"github.com/SevereCloud/vksdk/v2/api"
	"github.com/SevereCloud/vksdk/v2/object"
	"github.com/sirupsen/logrus"
)

var packageLogFields = logrus.Fields{
	"package": "service",
}

type MediaContentCourier struct {
	communityVkApi     *api.VK
	userVkApi          *api.VK
	phrasesRepo        repository.PhraseRepository
	contentCommandRepo repository.ContentCommandRepository
	listeningChannel   chan *botobject.ContentRequestCommand
	commandCollector   map[string]content.AttachmentsContentCollector
}

func NewMediaContentCourier(
	communityVkApi,
	userVkApi *api.VK,
	phrasesRepo repository.PhraseRepository,
	contentCommandRepo repository.ContentCommandRepository,
	listeningChannel chan *botobject.ContentRequestCommand,
) *MediaContentCourier {
	return &MediaContentCourier{
		communityVkApi:     communityVkApi,
		userVkApi:          userVkApi,
		phrasesRepo:        phrasesRepo,
		contentCommandRepo: contentCommandRepo,
		listeningChannel:   listeningChannel,
		commandCollector:   make(map[string]content.AttachmentsContentCollector),
	}
}

func (courier *MediaContentCourier) ReceiveAndDeliver() {
	for {
		select {
		case request := <-courier.listeningChannel:
			user, err := vk.GetUserInfo(courier.communityVkApi, request.Event.UserID)
			if err != nil {
				logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
					"struct": "MediaContentCourier",
					"func":   "ReceiveAndDeliver",
					"err":    err,
					"user":   request.Event.UserID,
				}).Error("get user info error")
				continue
			}

			messageToSend := vk.BuildMessageUsingPersonalizedPhrase(
				request.Event.PeerID,
				user,
				model.ContentRequestType,
				courier.phrasesRepo.FindAllByType(model.ContentRequestType),
			)

			if collector := courier.commandCollector[request.Command.Name]; collector == nil {
				courier.commandCollector[request.Command.Name] = NewCachedRandomAttachmentsContentCollector(
					courier.userVkApi,
					request.GetAttachmentsType(),
					request.Command.Name,
					courier.contentCommandRepo,
					courier.getMaxCachedAttachments(request.GetAttachmentsType()),
					courier.getCacheRefreshThreshold(request.GetAttachmentsType()),
				)
			}

			mediaContent := courier.commandCollector[request.Command.Name].CollectOne()
			if len(mediaContent.Type) == 0 {
				logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
					"func":   "ReceiveAndDeliver",
					"struct": "MediaContentCourier",
				}).Warn("collected empty media content ignored")
				continue
			}

			messageToSend["attachment"] = courier.resolveContentID(mediaContent, request.GetAttachmentsType())
			_, err = courier.communityVkApi.MessagesSend(messageToSend)
			if _, messageContainsPhrase := messageToSend["message"]; !messageContainsPhrase {
				logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
					"func":   "ReceiveAndDeliver",
					"struct": "MediaContentCourier",
					"err":    err,
				}).Error("message send error")
			}
		}
	}
}

func (courier *MediaContentCourier) getMaxCachedAttachments(mediaType vk.AttachmentsType) int {
	switch mediaType {
	case vk.PhotoType:
		pictureMaxCachedAttachments, err := strconv.ParseInt(utils.GetEnvOrDefault(configs.ContentPictureMaxCachedAttachments), 10, 32)
		if err != nil {
			logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
				"func": "MediaContentCourier",
				"err":  err,
				"key":  configs.ContentPictureMaxCachedAttachments.Key,
			}).Fatal("parsing of env variable is failed")
		}

		return int(pictureMaxCachedAttachments)
	case vk.AudioType:
		audioMaxCachedAttachments, err := strconv.ParseInt(utils.GetEnvOrDefault(configs.ContentAudioMaxCachedAttachments), 10, 32)
		if err != nil {
			logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
				"func": "MediaContentCourier",
				"err":  err,
				"key":  configs.ContentAudioMaxCachedAttachments.Key,
			}).Fatal("parsing of env variable is failed")
		}

		return int(audioMaxCachedAttachments)
	}

	return 0
}

func (courier *MediaContentCourier) getCacheRefreshThreshold(mediaType vk.AttachmentsType) float32 {
	switch mediaType {
	case vk.PhotoType:
		pictureCacheRefreshThreshold, err := strconv.ParseFloat(utils.GetEnvOrDefault(configs.ContentPictureCacheRefreshThreshold), 32)
		if err != nil {
			logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
				"func": "MediaContentCourier",
				"err":  err,
				"key":  configs.ContentPictureCacheRefreshThreshold.Key,
			}).Fatal("parsing of env variable is failed")
		}

		return float32(pictureCacheRefreshThreshold)
	case vk.AudioType:
		audioCacheRefreshThreshold, err := strconv.ParseFloat(utils.GetEnvOrDefault(configs.ContentAudioCacheRefreshThreshold), 32)
		if err != nil {
			logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
				"func": "MediaContentCourier",
				"err":  err,
				"key":  configs.ContentAudioCacheRefreshThreshold.Key,
			}).Fatal("parsing of env variable is failed")
		}

		return float32(audioCacheRefreshThreshold)
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
	}

	return ""
}
