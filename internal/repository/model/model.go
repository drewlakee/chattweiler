package model

import (
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

// CsvContentCommand storage specific object of ContentCommand
type CsvContentCommand struct {
	ID                int    `csv:"id"`
	Commands          string `csv:"commands"`
	MediaContentTypes string `csv:"media_types"`
	CommunityIDs      string `csv:"community_ids"`
}

// ContentCommand domain object
type ContentCommand struct {
	id int

	// aliases to call for request (e.g. "sing", "song" are refer to one command)
	commandsList []string

	// media content type which command supposed to deliver on call
	mediaContentType []MediaContentType

	// communities that are able for command to use as content sources
	communityIDsList []string
}

func (contentCommand ContentCommand) GetAliases() []string {
	return contentCommand.commandsList
}

func (contentCommand ContentCommand) GetMediaTypes() []MediaContentType {
	return contentCommand.mediaContentType
}

func (contentCommand ContentCommand) GetID() int {
	return contentCommand.id
}

func (contentCommand ContentCommand) GetCommunityIDs() []string {
	return contentCommand.communityIDsList
}

func NewContentCommand(
	id int,
	commands []string,
	mediaContentType []MediaContentType,
	communityIDs []string,
) ContentCommand {
	return ContentCommand{
		id:               id,
		commandsList:     commands,
		mediaContentType: mediaContentType,
		communityIDsList: communityIDs,
	}
}
