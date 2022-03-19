package messages

import (
	"chattweiler/pkg/app/utils"
	"chattweiler/pkg/repository/model"
	"chattweiler/pkg/roulette"
	"fmt"
	"github.com/SevereCloud/vksdk/v2/api"
	"github.com/SevereCloud/vksdk/v2/api/params"
	"github.com/SevereCloud/vksdk/v2/object"
	"github.com/sirupsen/logrus"
	"strconv"
	"strings"
)

var packageLogFields = logrus.Fields{
	"package": "messages",
}

func BuildMessageUsingPersonalizedPhrase(peerId int, user *object.UsersUser, phrases []model.Phrase) api.Params {
	phrase := roulette.Spin(phrases...)
	if phrase == nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"func":     "BuildMessageUsingPersonalizedPhrase",
			"fallback": "empty api params",
		}).Warn("empty phrases passed")
		return api.Params{}
	}

	builder := params.NewMessagesSendBuilder()
	builder.PeerID(peerId)
	builder.RandomID(0)

	useFirstNameInsteadUsername, err := strconv.ParseBool(utils.GetEnvOrDefault("chat.use.first.name.instead.username", "false"))
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"func": "BuildMessageUsingPersonalizedPhrase",
		}).Error("chat.use.first.name.instead.username parse error")
	}

	if phrase.IsUserTemplated {
		if useFirstNameInsteadUsername {
			builder.Message(strings.ReplaceAll(phrase.Text, "%username%", fmt.Sprintf("@%s (%s)", user.ScreenName, user.FirstName)))
		} else {
			builder.Message(strings.ReplaceAll(phrase.Text, "%username%", "@"+user.ScreenName))
		}
	} else {
		builder.Message(fmt.Sprintf("%s, \n\n%s", "@"+user.ScreenName, phrase.Text))
	}

	if phrase.IsAudioAccompaniment {
		if phrase.VkAudioId.Valid {
			builder.Attachment(phrase.VkAudioId.String)
		} else {
			logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
				"func":      "BuildMessageUsingPersonalizedPhrase",
				"phrase_id": phrase.PhraseID,
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
	builder.Message(phrase.Text)

	if phrase.IsAudioAccompaniment {
		if phrase.VkAudioId.Valid {
			builder.Attachment(phrase.VkAudioId.String)
		} else {
			logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
				"func":      "BuildMessagePhrase",
				"phrase_id": phrase.PhraseID,
			}).Warn("pharse specified with audio accompaniment, but audio_id doesn't pointed")
		}
	}

	return builder.Params
}
