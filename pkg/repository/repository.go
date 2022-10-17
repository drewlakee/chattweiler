// Package repository provides interfaces for data storages
package repository

import (
	"chattweiler/pkg/repository/model"
)

var logPackage = "repository"

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
