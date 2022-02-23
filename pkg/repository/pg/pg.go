package pg

import (
	"chattweiler/pkg/repository/model"
	"fmt"
	"github.com/jmoiron/sqlx"
)

type CachedPgPhraseRepository struct {
	db      *sqlx.DB
	phrases []model.Phrase
}

func NewCachedPgPhraseRepository(db *sqlx.DB) *CachedPgPhraseRepository {
	return &CachedPgPhraseRepository{db, nil}
}

func (cachedPgPhraseRepository *CachedPgPhraseRepository) FindAll() []model.Phrase {
	if cachedPgPhraseRepository.phrases != nil && len(cachedPgPhraseRepository.phrases) != 0 {
		return cachedPgPhraseRepository.phrases
	}

	query :=
		"SELECT phrase_id, weight, pt.name AS phrase_type, is_user_templated, is_audio_accompaniment, vk_audio_id, text" +
			"FROM phrase AS p, phrase_type AS pt" +
			"WHERE p.type = pt.type_id"

	err := cachedPgPhraseRepository.db.Select(&cachedPgPhraseRepository.phrases, query, "")
	if err != nil {
		fmt.Println(err)
		return []model.Phrase{}
	}

	return cachedPgPhraseRepository.phrases
}

type PgMembershipWarningRepository struct {
	db *sqlx.DB
}

func NewPgMembershipWarningRepository(db *sqlx.DB) *PgMembershipWarningRepository {
	return &PgMembershipWarningRepository{db}
}

func (pgMembershipWarningRepository *PgMembershipWarningRepository) InsertAll(warnings ...model.MembershipWarning) bool {
	insert :=
		"INSERT INTO membership_warning (user_id, username, first_warning_ts, grace_period, is_relevant)" +
			"VALUES ($1, $2, $3, $4, $5)"

	beginTx := pgMembershipWarningRepository.db.MustBegin()
	for _, warning := range warnings {
		pgMembershipWarningRepository.db.MustExec(
			insert,
			warning.UserID,
			warning.Username,
			warning.FirstWarningTs,
			warning.GracePeriod,
			warning.IsRelevant,
		)
	}

	err := beginTx.Commit()
	if err != nil {
		fmt.Println(err)
		return false
	}

	return true
}

func (pgMembershipWarningRepository *PgMembershipWarningRepository) FindAll() []model.MembershipWarning {
	query :=
		"SELECT warning_id, user_id, username, first_warning_ts, grace_period, is_relevant" +
			"FROM membership_warning"

	var warnings []model.MembershipWarning
	err := pgMembershipWarningRepository.db.Select(&warnings, query, "")
	if err != nil {
		fmt.Println(err)
		return []model.MembershipWarning{}
	}

	return warnings
}

func (pgMembershipWarningRepository *PgMembershipWarningRepository) FindAllRelevant() []model.MembershipWarning {
	query :=
		"SELECT warning_id, user_id, username, first_warning_ts, grace_period, is_relevant" +
			"FROM membership_warning" +
			"WHERE is_relevant = true"

	var warnings []model.MembershipWarning
	err := pgMembershipWarningRepository.db.Select(&warnings, query, "")
	if err != nil {
		fmt.Println(err)
		return []model.MembershipWarning{}
	}

	return warnings
}
