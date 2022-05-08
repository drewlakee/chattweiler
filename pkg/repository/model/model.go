package model

import (
	"chattweiler/pkg/repository/model/types"
	"database/sql"
	"time"
)

type Phrase struct {
	PhraseID             int              `db:"phrase_id" csv:"phrase_id"`
	Weight               int              `db:"weight" csv:"weight"`
	PhraseType           types.PhraseType `db:"phrase_type" csv:"phrase_type"`
	IsUserTemplated      bool             `db:"is_user_templated" csv:"is_user_templated"`
	IsAudioAccompaniment bool             `db:"is_audio_accompaniment" csv:"is_audio_accompaniment"`
	VkAudioId            sql.NullString   `db:"vk_audio_id" csv:"vk_audio_id"`
	Text                 string           `db:"text" csv:"text"`
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
	SourceID      int                     `db:"source_id" csv:"source_id"`
	VkCommunityID string                  `db:"vk_community_id" csv:"vk_community_id"`
	SourceType    types.ContentSourceType `db:"source_type" csv:"source_type"`
}
