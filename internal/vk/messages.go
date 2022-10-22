package vk

import (
	"chattweiler/internal/configs"
	"chattweiler/internal/logging"
	"chattweiler/internal/repository/model"
	"chattweiler/internal/roulette"
	"chattweiler/internal/utils"
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	"github.com/SevereCloud/vksdk/v2/api"
	"github.com/SevereCloud/vksdk/v2/api/params"
	"github.com/SevereCloud/vksdk/v2/object"
)

func BuildDirectedMessage(peerId int) api.Params {
	builder := params.NewMessagesSendBuilder()
	builder.PeerID(peerId)
	builder.RandomID(rand.Int())
	return builder.Params
}

func BuildMessageUsingPersonalizedPhrase(
	peerId int,
	user *object.UsersUser,
	phrases []model.Phrase,
) api.Params {
	phrase := roulette.Spin(phrases...)
	builder := params.NewMessagesSendBuilder()
	builder.PeerID(peerId)
	builder.RandomID(rand.Int())

	useFirstNameInsteadUsername, err := strconv.ParseBool(utils.GetEnvOrDefault(configs.ChatUseFirstNameInsteadUsername))
	if err != nil {
		logging.Log.Error(logPackage, "BuildMessageUsingPersonalizedPhrase", err, "%s: parsing of env variable is failed", configs.ChatUseFirstNameInsteadUsername.Key)
	}

	if phrase.UserTemplated() {
		if useFirstNameInsteadUsername {
			builder.Message(strings.ReplaceAll(phrase.GetText(), "%username%", fmt.Sprintf("@%s (%s)", user.ScreenName, user.FirstName)))
		} else {
			builder.Message(strings.ReplaceAll(phrase.GetText(), "%username%", "@"+user.ScreenName))
		}
	} else {
		builder.Message(fmt.Sprintf("%s, \n\n%s", "@"+user.ScreenName, phrase.GetText()))
	}

	appendAttachments(phrase, builder)
	return builder.Params
}

func BuildMessageWithRandomPhrase(peerId int, phrases []model.Phrase) api.Params {
	phrase := roulette.Spin(phrases...)
	builder := params.NewMessagesSendBuilder()
	builder.PeerID(peerId)
	builder.RandomID(0)
	builder.Message(phrase.GetText())
	appendAttachments(phrase, builder)
	return builder.Params
}

func appendAttachments(phrase model.Phrase, builder *params.MessagesSendBuilder) {
	var attachments []string

	if phrase.HasAudioAccompaniment() {
		attachments = append(attachments, phrase.GetVkAudioId())
	}

	if phrase.HasGifAccompaniment() {
		attachments = append(attachments, phrase.GetVkGifId())
	}

	if len(attachments) != 0 {
		builder.Attachment(strings.Join(attachments, ","))
	}
}
