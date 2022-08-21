package vk

import (
	"chattweiler/pkg/configs"
	"chattweiler/pkg/repository/model"
	"chattweiler/pkg/roulette"
	"chattweiler/pkg/utils"
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	"github.com/SevereCloud/vksdk/v2/api"
	"github.com/SevereCloud/vksdk/v2/api/params"
	"github.com/SevereCloud/vksdk/v2/object"
	"github.com/sirupsen/logrus"
)

func BuildMessageUsingPersonalizedPhrase(
	peerId int,
	user *object.UsersUser,
	phrasesType model.PhraseType,
	phrases []model.Phrase,
) api.Params {
	phrase := roulette.Spin(phrases...)
	if phrase == nil {
		if suppress, _ := strconv.ParseBool(utils.GetEnvOrDefault(configs.PhrasesSuppressLogsMissedPhrases)); !suppress {
			logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
				"func":       "BuildMessageUsingPersonalizedPhrase",
				"phraseType": phrasesType,
			}).Warn("were passed empty phrases, but response message supposed to be with a phrase")
		}
	}

	builder := params.NewMessagesSendBuilder()
	builder.PeerID(peerId)
	builder.RandomID(rand.Int())

	if phrase == nil {
		builder.Message(" ")
		return builder.Params
	}

	useFirstNameInsteadUsername, err := strconv.ParseBool(utils.GetEnvOrDefault(configs.ChatUseFirstNameInsteadUsername))
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"func": "BuildMessageUsingPersonalizedPhrase",
			"key":  configs.ChatUseFirstNameInsteadUsername.Key,
		}).Error("parsing of env variable is failed")
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

func BuildMessagePhrase(peerId int, phrases []model.Phrase) api.Params {
	phrase := roulette.Spin(phrases...)

	if phrase == nil {
		if suppress, _ := strconv.ParseBool(utils.GetEnvOrDefault(configs.PhrasesSuppressLogsMissedPhrases)); !suppress {
			logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
				"func":     "BuildMessagePhrase",
				"fallback": "empty api params",
			}).Warn("empty phrases were passed")
		}
	}

	builder := params.NewMessagesSendBuilder()
	builder.PeerID(peerId)
	builder.RandomID(0)

	if phrase == nil {
		builder.Message("")
		return builder.Params
	}

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
