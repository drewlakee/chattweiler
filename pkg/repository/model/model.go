package model

import (
	"database/sql"
	"time"
)

type Phrase interface {
	GetID() int
	GetWeight() int
	GetPhraseType() PhraseType
	UserTemplated() bool
	HasAudioAccompaniment() bool
	GetVkAudioId() string
	GetText() string

	NullableVkAudio() bool
}

type PhrasePg struct {
	PhraseID             int            `db:"phrase_id" csv:"phrase_id"`
	Weight               int            `db:"weight" csv:"weight"`
	PhraseType           PhraseType     `db:"phrase_type" csv:"phrase_type"`
	IsUserTemplated      bool           `db:"is_user_templated" csv:"is_user_templated"`
	IsAudioAccompaniment bool           `db:"is_audio_accompaniment" csv:"is_audio_accompaniment"`
	VkAudioId            sql.NullString `db:"vk_audio_id" csv:"vk_audio_id"`
	Text                 string         `db:"text" csv:"text"`
}

func (p PhrasePg) GetID() int {
	return p.PhraseID
}

func (p PhrasePg) GetWeight() int {
	return p.Weight
}

func (p PhrasePg) GetPhraseType() PhraseType {
	return p.PhraseType
}

func (p PhrasePg) UserTemplated() bool {
	return p.IsUserTemplated
}

func (p PhrasePg) HasAudioAccompaniment() bool {
	return p.IsAudioAccompaniment
}

func (p PhrasePg) GetVkAudioId() string {
	return p.VkAudioId.String
}

func (p PhrasePg) GetText() string {
	return p.Text
}

func (p PhrasePg) NullableVkAudio() bool {
	return p.VkAudioId.Valid
}

type PhraseCsv struct {
	PhraseID             int        `csv:"phrase_id"`
	Weight               int        `csv:"weight"`
	PhraseType           PhraseType `csv:"phrase_type"`
	IsUserTemplated      bool       `csv:"is_user_templated"`
	IsAudioAccompaniment bool       `csv:"is_audio_accompaniment"`
	VkAudioId            string     `csv:"vk_audio_id"`
	Text                 string     `csv:"text"`
}

func (p PhraseCsv) GetID() int {
	return p.PhraseID
}

func (p PhraseCsv) GetWeight() int {
	return p.Weight
}

func (p PhraseCsv) GetPhraseType() PhraseType {
	return p.PhraseType
}

func (p PhraseCsv) UserTemplated() bool {
	return p.IsUserTemplated
}

func (p PhraseCsv) HasAudioAccompaniment() bool {
	return p.IsAudioAccompaniment
}

func (p PhraseCsv) GetVkAudioId() string {
	return p.VkAudioId
}

func (p PhraseCsv) GetText() string {
	return p.Text
}

func (p PhraseCsv) NullableVkAudio() bool {
	return p.VkAudioId == ""
}

type MembershipWarning struct {
	WarningID      int           `db:"warning_id" csv:"warning_id"`
	UserID         int           `db:"user_id" csv:"user_id"`
	Username       string        `db:"username" csv:"username"`
	FirstWarningTs time.Time     `db:"first_warning_ts" csv:"first_warning_ts"`
	GracePeriod    time.Duration `db:"grace_period_ns" csv:"grace_period_ns"`
	IsRelevant     bool          `db:"is_relevant" csv:"is_relevant"`
}

type ContentSource struct {
	SourceID      int               `db:"source_id" csv:"source_id"`
	VkCommunityID string            `db:"vk_community_id" csv:"vk_community_id"`
	SourceType    ContentSourceType `db:"source_type" csv:"source_type"`
}
