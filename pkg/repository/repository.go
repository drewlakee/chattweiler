package repository

import "chattweiler/pkg/repository/model"

type PhraseRepository interface {
	FindAll() []model.Phrase
}

type MembershipWarningRepository interface {
	InsertAll(warnings ...model.MembershipWarning) bool
	FindAll() []model.MembershipWarning
	FindAllRelevant() []model.MembershipWarning
}
