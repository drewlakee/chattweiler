// Package model provides objects for repositories
package model

import (
	"database/sql"
	"strings"
	"time"
)

type Phrase interface {
	GetID() int
	GetWeight() int
	GetPhraseType() PhraseType
	UserTemplated() bool
	HasAudioAccompaniment() bool
	GetVkAudioId() string
	HasGifAccompaniment() bool
	GetVkGifId() string
	GetText() string
}

type PhrasePg struct {
	PhraseID   int            `db:"phrase_id"`
	Weight     int            `db:"weight"`
	PhraseType PhraseType     `db:"phrase_type"`
	VkAudioId  sql.NullString `db:"vk_audio_id"`
	VkGifId    sql.NullString `db:"vk_gif_id"`
	Text       string         `db:"text"`
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
	return strings.Contains(p.Text, "%username%")
}

func (p PhrasePg) HasAudioAccompaniment() bool {
	return p.VkAudioId.Valid
}

func (p PhrasePg) GetVkAudioId() string {
	return p.VkAudioId.String
}

func (p PhrasePg) GetText() string {
	return p.Text
}

func (p PhrasePg) HasGifAccompaniment() bool {
	return p.VkGifId.Valid
}

func (p PhrasePg) GetVkGifId() string {
	return p.VkGifId.String
}

type PhraseCsv struct {
	PhraseID   int        `csv:"phrase_id"`
	Weight     int        `csv:"weight"`
	PhraseType PhraseType `csv:"phrase_type"`
	VkAudioId  string     `csv:"vk_audio_id"`
	VkGifId    string     `csv:"vk_gif_id"`
	Text       string     `csv:"text"`
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
	return strings.Contains(p.Text, "%username%")
}

func (p PhraseCsv) HasAudioAccompaniment() bool {
	id := strings.TrimSpace(p.VkAudioId)
	return id != "" && !strings.EqualFold(id, "null")
}

func (p PhraseCsv) GetVkAudioId() string {
	return p.VkAudioId
}

func (p PhraseCsv) GetText() string {
	return p.Text
}

func (p PhraseCsv) HasGifAccompaniment() bool {
	id := strings.TrimSpace(p.VkGifId)
	return id != "" && !strings.EqualFold(id, "null")
}

func (p PhraseCsv) GetVkGifId() string {
	return p.VkGifId
}

type MembershipWarning struct {
	WarningID      int       `db:"warning_id" csv:"warning_id"`
	UserID         int       `db:"user_id" csv:"user_id"`
	Username       string    `db:"username" csv:"username"`
	FirstWarningTs time.Time `db:"first_warning_ts" csv:"first_warning_ts"`
	GracePeriod    string    `db:"grace_period" csv:"grace_period"`
	IsRelevant     bool      `db:"is_relevant" csv:"is_relevant"`
}

type ContentCommand struct {
	ID               int              `db:"id" csv:"id"`
	command          string           `db:"command" csv:"command"`
	MediaContentType MediaContentType `db:"media_type" csv:"media_type"`
	vkCommunityID    string           `db:"community_id" csv:"community_id"`
}

func (contentCommand *ContentCommand) GetCommunityIDs() []string {
	return strings.Split(contentCommand.vkCommunityID, ",")
}

func (contentCommand ContentCommand) GetAliases() []string {
	return strings.Split(contentCommand.command, ",")
}

func (contentCommand ContentCommand) ContainsAlias(alias string) bool {
	for _, existingAlias := range contentCommand.GetAliases() {
		if strings.EqualFold(existingAlias, alias) {
			return true
		}
	}
	return false
}
