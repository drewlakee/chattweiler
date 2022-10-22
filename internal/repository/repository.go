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

type ContentCommandRepository interface {
	FindAll() []model.ContentCommand
	FindByCommandAlias(command string) *model.ContentCommand
	FindById(ID int) *model.ContentCommand
}
