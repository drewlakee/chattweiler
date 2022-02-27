package repository

import (
	"chattweiler/pkg/repository/model"
	"chattweiler/pkg/repository/model/types"
)

type PhraseRepository interface {
	FindAll() []model.Phrase
	FindAllByType(phraseType types.PhraseType) []model.Phrase
}

type MembershipWarningRepository interface {
	Insert(model.MembershipWarning) bool
	UpdateAllToUnRelevant(...model.MembershipWarning) bool
	FindAll() []model.MembershipWarning
	FindAllRelevant() []model.MembershipWarning
}
