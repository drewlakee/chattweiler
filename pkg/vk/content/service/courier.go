package service

import (
	botobject "chattweiler/pkg/bot/object"
	"chattweiler/pkg/repository"
	"chattweiler/pkg/repository/model"
	"chattweiler/pkg/vk"
	"chattweiler/pkg/vk/content"

	"github.com/SevereCloud/vksdk/v2/api"
	"github.com/SevereCloud/vksdk/v2/object"
	"github.com/sirupsen/logrus"
)

var packageLogFields = logrus.Fields{
	"package": "service",
}

type MediaContentCourier struct {
	communityVkApi              *api.VK
	userVkApi                   *api.VK
	attachmentsContentCollector content.AttachmentsContentCollector
	phrasesRepo                 repository.PhraseRepository
}

func NewMediaContentCourier(
	communityVkApi,
	userVkApi *api.VK,
	attachmentsContentCollector content.AttachmentsContentCollector,
	phrasesRepo repository.PhraseRepository,
) *MediaContentCourier {
	return &MediaContentCourier{
		communityVkApi:              communityVkApi,
		userVkApi:                   userVkApi,
		attachmentsContentCollector: attachmentsContentCollector,
		phrasesRepo:                 phrasesRepo,
	}
}

func (courier *MediaContentCourier) ReceiveAndDeliver(
	deliverPhraseType model.PhraseType,
	contentRequestChannel <-chan *botobject.ContentRequestCommand,
) {
	for {
		select {
		case <-contentRequestChannel:
			return
		case requestMessage := <-contentRequestChannel:
			user, err := vk.GetUserInfo(courier.communityVkApi, requestMessage.RequestEvent.UserID)
			if err != nil {
				logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
					"struct": "MediaContentCourier",
					"func":   "ReceiveAndDeliver",
					"err":    err,
					"user":   requestMessage.RequestEvent.UserID,
				}).Error("get user info error")
				continue
			}

			messageToSend := vk.BuildMessageUsingPersonalizedPhrase(
				requestMessage.RequestEvent.PeerID,
				user,
				deliverPhraseType,
				courier.phrasesRepo.FindAllByType(deliverPhraseType),
			)

			mediaContent := courier.attachmentsContentCollector.CollectOne()

			if len(mediaContent.Type) == 0 {
				logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
					"func":   "ReceiveAndDeliver",
					"struct": "MediaContentCourier",
				}).Warn("collected empty media content ignored")
				continue
			}

			messageToSend["attachment"] = courier.resolveContentID(mediaContent, requestMessage.Type)
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
