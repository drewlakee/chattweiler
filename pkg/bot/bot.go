package bot

import (
	"chattweiler/pkg/repository"
	"github.com/SevereCloud/vksdk/v2/api"
	vklp "github.com/SevereCloud/vksdk/v2/longpoll-user"
	vklpwrapper "github.com/SevereCloud/vksdk/v2/longpoll-user/v3"
	wrapper "github.com/SevereCloud/vksdk/v2/longpoll-user/v3"
)

type Bot struct {
	vkapi                  *api.VK
	vklp                   *vklp.LongPoll
	vklpwrapper            *wrapper.Wrapper
	phraseRepo             *repository.PhraseRepository
	membershipWarningsRepo *repository.MembershipWarningRepository
}

func NewBot(vkToken string, phraseRepo *repository.PhraseRepository, membershipWarningsRepo *repository.MembershipWarningRepository) *Bot {
	vkapi := api.NewVK(vkToken)

	lp, err := vklp.NewLongPoll(vkapi, 0)
	if err != nil {
		panic(err)
	}

	wrappedlp := vklpwrapper.NewWrapper(lp)

	return &Bot{
		vkapi:                  vkapi,
		vklp:                   lp,
		vklpwrapper:            wrappedlp,
		phraseRepo:             phraseRepo,
		membershipWarningsRepo: membershipWarningsRepo,
	}
}

func (bot *Bot) Start() error {
	err := bot.vklp.Run()
	if err != nil {
		return err
	}

	return nil
}
