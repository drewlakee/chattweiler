package model

import (
	"strings"
	"time"
)

type Phrase struct {
	PhraseID   int        `csv:"phrase_id"`
	Weight     int        `csv:"weight"`
	PhraseType PhraseType `csv:"phrase_type"`
	VkAudioId  string     `csv:"vk_audio_id"`
	VkGifId    string     `csv:"vk_gif_id"`
	Text       string     `csv:"text"`
}

func (p Phrase) UserTemplated() bool {
	return strings.Contains(p.Text, "%username%")
}

func (p Phrase) HasAudioAccompaniment() bool {
	id := strings.TrimSpace(p.VkAudioId)
	return id != "" && !strings.EqualFold(id, "null")
}

func (p Phrase) HasGifAccompaniment() bool {
	id := strings.TrimSpace(p.VkGifId)
	return id != "" && !strings.EqualFold(id, "null")
}

type MembershipWarning struct {
	WarningID      int       `db:"warning_id" csv:"warning_id"`
	UserID         int       `db:"user_id" csv:"user_id"`
	Username       string    `db:"username" csv:"username"`
	FirstWarningTs time.Time `db:"first_warning_ts" csv:"first_warning_ts"`
	GracePeriod    string    `db:"grace_period" csv:"grace_period"`
	IsRelevant     bool      `db:"is_relevant" csv:"is_relevant"`
}

// CsvCommand storage specific object of Command
type CsvCommand struct {
	ID                int         `csv:"id"`
	Commands          string      `csv:"commands"`
	Type              CommandType `csv:"command_type"`
	MediaContentTypes string      `csv:"media_types"`
	CommunityIDs      string      `csv:"community_ids"`
}

// Command domain object
type Command struct {
	ID int

	Type CommandType

	// aliases to call for request (e.g. "sing", "song" are refer to one command)
	Aliases []string

	ContentDescriptor ContentDescriptor
}

type ContentDescriptor struct {
	// media content type which command supposed to deliver on call
	MediaContentType []MediaContentType

	// communities that are able for command to use as content sources
	CommunitySourceIDs []string
}

func NewCommand(
	id int,
	commandType CommandType,
	commands []string,
	mediaContentType []MediaContentType,
	communityIDs []string,
) Command {
	var command Command

	command.ID = id
	command.Aliases = commands
	command.Type = commandType

	switch commandType {
	case ContentCommand:
		command.ContentDescriptor = ContentDescriptor{
			MediaContentType:   mediaContentType,
			CommunitySourceIDs: communityIDs,
		}
	}

	return command
}
