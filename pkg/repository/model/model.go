package model

import (
	"chattweiler/pkg/repository/model/types"
	"database/sql"
	"time"
)

type Phrase struct {
	PhraseID             int              `db:"phrase_id"`
	Weight               int              `db:"weight"`
	PhraseType           types.PhraseType `db:"phrase_type"`
	IsUserTemplated      bool             `db:"is_user_templated"`
	IsAudioAccompaniment bool             `db:"is_audio_accompaniment"`
	VkAudioId            sql.NullString   `db:"vk_audio_id"`
	Text                 string           `db:"text"`
}

type MembershipWarning struct {
	WarningID      int           `db:"warning_id"`
	UserID         int           `db:"user_id"`
	Username       string        `db:"username"`
	FirstWarningTs time.Time     `db:"first_warning_ts"`
	GracePeriod    time.Duration `db:"grace_period_ns"`
	IsRelevant     bool          `db:"is_relevant"`
}

type ContentSource struct {
	SourceID      int                     `db:"source_id"`
	VkCommunityID string                  `db:"vk_community_id"`
	SourceType    types.ContentSourceType `db:"source_type"`
}
