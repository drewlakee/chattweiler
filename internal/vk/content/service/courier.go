package service

import (
	botobject "chattweiler/internal/bot/object"
	"chattweiler/internal/logging"
	"chattweiler/internal/repository"
	"chattweiler/internal/repository/model"
	"chattweiler/internal/vk"
	"chattweiler/internal/vk/content"
	"time"

	"github.com/SevereCloud/vksdk/v2/api"
	"github.com/SevereCloud/vksdk/v2/object"
)

type MediaContentCourier struct {
	communityVkApi          *api.VK
	userVkApi               *api.VK
	phrasesRepo             repository.PhraseRepository
	contentCommandRepo      repository.CommandsRepository
	listeningChannel        chan *botobject.ContentRequestCommand
	commandCollectors       map[int]content.AttachmentsContentCollector
	garbageCleaningInterval time.Duration
	lastTsGarbageCollected  time.Time
}

func NewMediaContentCourier(
	communityVkApi,
	userVkApi *api.VK,
	phrasesRepo repository.PhraseRepository,
	contentCommandRepo repository.CommandsRepository,
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

			if _, alreadyExists := courier.commandCollectors[received.Command.ID]; !alreadyExists {
				courier.createNewCollectorForCommand(received)
			}

			if courier.lastTsGarbageCollected.Add(courier.garbageCleaningInterval).Before(time.Now()) {
				courier.removeGarbageCollectors()
			}

			mediaAttachment := courier.commandCollectors[received.Command.ID].CollectOne()
			if mediaAttachment == nil || len(mediaAttachment.Type) == 0 {
				logging.Log.Warn(logPackage, "MediaContentCourier.ReceiveAndDeliver", "collected empty media content ignored")
				courier.askToRetryRequest(received, user)
				continue
			}

			courier.deliverContentResponse(received, user, mediaAttachment)
		}
	}
}

func (courier *MediaContentCourier) deliverContentResponse(
	request *botobject.ContentRequestCommand,
	user *object.UsersUser,
	mediaContent *content.MediaAttachment,
) {
	phrases := courier.phrasesRepo.FindAllByType(model.ContentRequestType)

	var messageToSend api.Params
	if len(phrases) == 0 {
		messageToSend = vk.BuildDirectedMessage(request.Event.PeerID)
	} else {
		messageToSend = vk.BuildMessageUsingPersonalizedPhrase(request.Event.PeerID, user, phrases)
	}

	messageToSend["attachment"] = courier.resolveAttachmentID(mediaContent)
	_, err := courier.communityVkApi.MessagesSend(messageToSend)
	if err != nil {
		logging.Log.Error(logPackage, "MediaContentCourier.deliverContentResponse", err, "message sending error. Sent params: %v", messageToSend)
		courier.askToRetryRequest(request, user)
	}
}

func (courier *MediaContentCourier) createNewCollectorForCommand(request *botobject.ContentRequestCommand) {
	courier.commandCollectors[request.Command.ID] = NewCachedRandomAttachmentsContentCollector(
		courier.userVkApi,
		request.GetAttachmentsTypes(),
		request.Command.ID,
		courier.contentCommandRepo,
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
		logging.Log.Error(logPackage, "MediaContentCourier.askToRetryRequest", err, "message sending error. Sent params: %v", messageToSend)
	}
}

func (courier *MediaContentCourier) resolveAttachmentID(mediaContent *content.MediaAttachment) string {
	switch mediaContent.Type {
	case vk.AudioType:
		return mediaContent.Data.Audio.ToAttachment()
	case vk.PhotoType:
		return mediaContent.Data.Photo.ToAttachment()
	case vk.VideoType:
		return mediaContent.Data.Video.ToAttachment()
	}

	return ""
}

func (courier *MediaContentCourier) removeGarbageCollectors() {
	relevantCommands := courier.contentCommandRepo.FindAll()
	relevantCommandsMap := make(map[int]bool, len(relevantCommands))
	for _, command := range relevantCommands {
		relevantCommandsMap[command.ID] = true
	}

	for commandID := range courier.commandCollectors {
		if _, exist := relevantCommandsMap[commandID]; !exist {
			delete(courier.commandCollectors, commandID)
		}
	}
}
