package vk

import (
	"chattweiler/pkg/configs"
	"chattweiler/pkg/repository/model"
	"chattweiler/pkg/roulette"
	"chattweiler/pkg/utils"
	"fmt"
	"strconv"
	"strings"

	"github.com/SevereCloud/vksdk/v2/api"
	"github.com/SevereCloud/vksdk/v2/api/params"
	"github.com/SevereCloud/vksdk/v2/object"
	"github.com/sirupsen/logrus"
)

func BuildMessageUsingPersonalizedPhrase(peerId int, user *object.UsersUser, phrasesType model.PhraseType, phrases []model.Phrase) api.Params {
	phrase := roulette.Spin(phrases...)
	if phrase == nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"func":       "BuildMessageUsingPersonalizedPhrase",
			"phraseType": phrasesType,
		}).Warn("were passed empty phrases, but a message supposed to be with a phrase")
	}

	builder := params.NewMessagesSendBuilder()
	builder.PeerID(peerId)
	builder.RandomID(0)

	if phrase == nil {
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

	if phrase.HasAudioAccompaniment() {
		if !phrase.NullableVkAudio() {
			builder.Attachment(phrase.GetVkAudioId())
		} else {
			logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
				"func":      "BuildMessageUsingPersonalizedPhrase",
				"phrase_id": phrase.GetID(),
			}).Warn("phrase is specified with audio accompaniment, but audio_id doesn't pointed")
		}
	}

	return builder.Params
}

func BuildMessagePhrase(peerId int, phrases []model.Phrase) api.Params {
	phrase := roulette.Spin(phrases...)
	if phrase == nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"func":     "BuildMessagePhrase",
			"fallback": "empty api params",
		}).Warn("empty phrases passed in")
		return api.Params{}
	}

	builder := params.NewMessagesSendBuilder()
	builder.PeerID(peerId)
	builder.RandomID(0)
	builder.Message(phrase.GetText())

	if phrase.HasAudioAccompaniment() {
		if !phrase.NullableVkAudio() {
			builder.Attachment(phrase.GetVkAudioId())
		} else {
			logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
				"func":      "BuildMessagePhrase",
				"phrase_id": phrase.GetID(),
			}).Warn("phrase must be with audio, but there is no audio attachment")
		}
	}

	return builder.Params
}
