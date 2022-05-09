package courier

import (
	"chattweiler/pkg/repository"
	"chattweiler/pkg/repository/model/types"
	"chattweiler/pkg/vk"
	"chattweiler/pkg/vk/content"
	"chattweiler/pkg/vk/messages"
	"github.com/SevereCloud/vksdk/v2/api"
	lpwrapper "github.com/SevereCloud/vksdk/v2/longpoll-user/v3"
	"github.com/SevereCloud/vksdk/v2/object"
	"github.com/sirupsen/logrus"
)

var packageLogFields = logrus.Fields{
	"package": "courier",
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

func (courier *MediaContentCourier) ReceiveAndDeliver(deliverPhraseType types.PhraseType, deliverContentType content.AttachmentsType, contentRequestChannel <-chan lpwrapper.NewMessage) {
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

			apiParams := messages.BuildMessageUsingPersonalizedPhrase(
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

func (courier *MediaContentCourier) resolveContentID(mediaContent object.WallWallpostAttachment, deliverContentType content.AttachmentsType) string {
	switch deliverContentType {
	case content.Audio:
		return mediaContent.Audio.ToAttachment()
	case content.Photo:
		return mediaContent.Photo.ToAttachment()
	}

	return ""
}
