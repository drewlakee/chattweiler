package repository

import (
	"chattweiler/pkg/repository/model"

	"github.com/sirupsen/logrus"
)

var packageLogFields = logrus.Fields{
	"package": "repository",
}

type PhraseRepository interface {
	FindAll() []model.Phrase
	FindAllByType(phraseType model.PhraseType) []model.Phrase
}

type MembershipWarningRepository interface {
	Insert(model.MembershipWarning) bool
	UpdateAllToUnRelevant(...model.MembershipWarning) bool
	FindAllRelevant() []model.MembershipWarning
}

type ContentSourceRepository interface {
	FindAll() []model.ContentSource
	FindAllByType(sourceType model.ContentSourceType) []model.ContentSource
}
