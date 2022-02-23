package model

import "time"

type PhraseType string

const (
	Welcome PhraseType = "welcome"
	Goodbye PhraseType = "goodbye"
)

type Phrase struct {
	PhraseID             int        `db:"phrase_id"`
	Weight               int        `db:"weight"`
	PhraseType           PhraseType `db:"phrase_type"`
	IsUserTemplated      bool       `db:"is_user_templated"`
	IsAudioAccompaniment bool       `db:"is_audio_accompaniment"`
	VkAudioId            string     `db:"vk_audio_id"`
	Text                 string     `db:"text"`
}

type MembershipWarning struct {
	WarningID      int           `db:"warning_id"`
	UserID         int           `db:"user_id"`
	Username       string        `db:"username"`
	FirstWarningTs time.Time     `db:"first_warning_ts"`
	GracePeriod    time.Duration `db:"grace_period"`
	IsRelevant     bool          `db:"is_relevant"`
}
