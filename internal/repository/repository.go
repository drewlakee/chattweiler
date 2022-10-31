package repository

import (
	"chattweiler/internal/repository/model"
)

type PhraseRepository interface {
	FindAll() []model.Phrase
	FindAllByType(phraseType model.PhraseType) []model.Phrase
}

type MembershipWarningRepository interface {
	Insert(model.MembershipWarning) bool
	UpdateAllToIrrelevant(...model.MembershipWarning) bool
	FindAllRelevant() []model.MembershipWarning
}

type CommandsRepository interface {
	FindAll() []model.Command
	FindByCommandAlias(command string) *model.Command
	FindById(ID int) *model.Command
}
