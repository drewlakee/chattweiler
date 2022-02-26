package messages

import (
	"chattweiler/pkg/repository/model"
	"chattweiler/pkg/roulette"
	"fmt"
	"github.com/SevereCloud/vksdk/v2/api"
	"github.com/SevereCloud/vksdk/v2/api/params"
	"github.com/SevereCloud/vksdk/v2/object"
	"strings"
)

func BuildMessageUsingPersonalizedPhrase(peerId int, user *object.UsersUser, phrases []model.Phrase) api.Params {
	phrase := roulette.Spin(phrases...)
	if phrase == nil {
		fmt.Printf("Choosed empty phrase for user\n")
		return api.Params{}
	}

	builder := params.NewMessagesSendBuilder()
	builder.PeerID(peerId)
	builder.RandomID(0)

	if phrase.IsUserTemplated {
		builder.Message(strings.ReplaceAll(phrase.Text, "%username%", "@"+user.ScreenName))
	} else {
		builder.Message(fmt.Sprintf("%s, \n\n%s", "@"+user.ScreenName, phrase.Text))
	}

	if phrase.IsAudioAccompaniment {
		if phrase.VkAudioId.Valid {
			builder.Attachment(phrase.VkAudioId.String)
		} else {
			fmt.Println("Pharse specified with audio accompaniment, but audio_id doesn't pointed. phrase_id:", phrase.PhraseID)
		}
	}

	return builder.Params
}

func BuildMessagePhrase(peerId int, phrases []model.Phrase) api.Params {
	phrase := roulette.Spin(phrases...)
	if phrase == nil {
		fmt.Printf("Choosed empty phrase\n")
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
			fmt.Println("Pharse specified with audio accompaniment, but audio_id doesn't pointed. phrase_id:", phrase.PhraseID)
		}
	}

	return builder.Params
}
