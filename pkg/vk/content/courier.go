package content

import (
	"chattweiler/pkg/repository"
	"chattweiler/pkg/repository/model"
	"chattweiler/pkg/vk"

	"github.com/SevereCloud/vksdk/v2/api"
	lpwrapper "github.com/SevereCloud/vksdk/v2/longpoll-user/v3"
	"github.com/SevereCloud/vksdk/v2/object"
	"github.com/sirupsen/logrus"
)

type MediaContentCourier struct {
	communityVkApi              *api.VK
	userVkApi                   *api.VK
	attachmentsContentCollector AttachmentsContentCollector
	phrasesRepo                 repository.PhraseRepository
}

func NewMediaContentCourier(
	communityVkApi,
	userVkApi *api.VK,
	attachmentsContentCollector AttachmentsContentCollector,
	phrasesRepo repository.PhraseRepository,
) *MediaContentCourier {
	return &MediaContentCourier{
		communityVkApi:              communityVkApi,
		userVkApi:                   userVkApi,
		attachmentsContentCollector: attachmentsContentCollector,
		phrasesRepo:                 phrasesRepo,
	}
}

func (courier *MediaContentCourier) ReceiveAndDeliver(deliverPhraseType model.PhraseType, deliverContentType AttachmentsType, contentRequestChannel <-chan lpwrapper.NewMessage) {
	for {
		select {
		case requestMessage := <-contentRequestChannel:
			user, err := vk.GetUserInfo(courier.communityVkApi, requestMessage.AdditionalData.From)
			if err != nil {
				logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
					"struct": "MediaContentCourier",
					"func":   "ReceiveAndDeliver",
					"err":    err,
					"user":   requestMessage.AdditionalData.From,
				}).Error("get user info error")
				continue
			}

			apiParams := vk.BuildMessageUsingPersonalizedPhrase(
				requestMessage.PeerID,
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

			apiParams["attachment"] = courier.resolveContentID(mediaContent, deliverContentType)
			_, err = courier.communityVkApi.MessagesSend(apiParams)
			if err != nil {
				logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
					"func":   "ReceiveAndDeliver",
					"struct": "MediaContentCourier",
					"err":    err,
				}).Error("message send error")
			}
		}
	}
}

func (courier *MediaContentCourier) resolveContentID(mediaContent object.WallWallpostAttachment, deliverContentType AttachmentsType) string {
	switch deliverContentType {
	case AudioType:
		return mediaContent.Audio.ToAttachment()
	case PhotoType:
		return mediaContent.Photo.ToAttachment()
	}

	return ""
}
