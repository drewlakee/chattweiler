package pg

import (
	"chattweiler/pkg/repository/model"
	"chattweiler/pkg/repository/model/types"
	"fmt"
	"github.com/jmoiron/sqlx"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

type CachedPgPhraseRepository struct {
	db                   *sqlx.DB
	cacheRefreshInterval time.Duration
	lastCacheRefresh     time.Time
	refreshMutex         sync.Mutex
	phrases              []model.Phrase
}

func NewCachedPgPhraseRepository(db *sqlx.DB, cacheRefreshInterval time.Duration) *CachedPgPhraseRepository {
	return &CachedPgPhraseRepository{
		db:                   db,
		cacheRefreshInterval: cacheRefreshInterval,
		lastCacheRefresh:     time.Now(),
		phrases:              nil,
	}
}

func (cachedPgPhraseRepository *CachedPgPhraseRepository) FindAll() []model.Phrase {
	if time.Now().Before(cachedPgPhraseRepository.lastCacheRefresh.Add(cachedPgPhraseRepository.cacheRefreshInterval)) {
		// atomic phrases read
		phrasesPtr := atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&cachedPgPhraseRepository.phrases)))
		if phrasesPtr != nil {
			phrases := *(*[]model.Phrase)(phrasesPtr)
			if len(phrases) != 0 {
				return phrases
			}
		}
	}

	// cache refresh lock
	cachedPgPhraseRepository.refreshMutex.Lock()
	defer cachedPgPhraseRepository.refreshMutex.Unlock()

	query :=
		"SELECT phrase_id, weight, pt.name AS phrase_type, is_user_templated, is_audio_accompaniment, vk_audio_id, text " +
			"FROM phrase AS p, phrase_type AS pt " +
			"WHERE p.type = pt.type_id "

	var updatedPhrases []model.Phrase
	err := cachedPgPhraseRepository.db.Select(&updatedPhrases, query)
	if err != nil {
		fmt.Printf("Error: %s, Query: %s\n", err.Error(), query)
		return []model.Phrase{}
	}

	// atomic phrases write
	updatedPhrasesPtr := unsafe.Pointer(&updatedPhrases)
	atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&cachedPgPhraseRepository.phrases)), updatedPhrasesPtr)

	cachedPgPhraseRepository.lastCacheRefresh = time.Now()
	return updatedPhrases
}

func (cachedPgPhraseRepository *CachedPgPhraseRepository) FindAllByType(phraseType types.PhraseType) []model.Phrase {
	var phrases []model.Phrase
	for _, phrase := range cachedPgPhraseRepository.FindAll() {
		if phraseType == phrase.PhraseType {
			phrases = append(phrases, phrase)
		}
	}
	return phrases
}

type PgMembershipWarningRepository struct {
	db *sqlx.DB
}

func NewPgMembershipWarningRepository(db *sqlx.DB) *PgMembershipWarningRepository {
	return &PgMembershipWarningRepository{db}
}

func (pgMembershipWarningRepository *PgMembershipWarningRepository) Insert(warning model.MembershipWarning) bool {
	insert :=
		"INSERT INTO membership_warning (user_id, username, first_warning_ts, grace_period_ns, is_relevant) " +
			"VALUES ($1, $2, $3, $4, $5)"

	_, err := pgMembershipWarningRepository.db.Exec(
		insert,
		warning.UserID,
		warning.Username,
		warning.FirstWarningTs,
		warning.GracePeriod,
		warning.IsRelevant,
	)

	if err != nil {
		fmt.Printf("Error: %s, Insert: %s, Params: %v\n", err.Error(), insert, warning)
		return false
	}

	return true
}

func (pgMembershipWarningRepository *PgMembershipWarningRepository) UpdateAllToUnRelevant(warnings ...model.MembershipWarning) bool {
	update :=
		"UPDATE membership_warning " +
			"SET is_relevant=false " +
			"WHERE warning_id = $1"

	tx, err := pgMembershipWarningRepository.db.Begin()
	if err != nil {
		fmt.Printf("Error: %s, Update: %s\n", err.Error(), update)
		return false
	}

	for _, warning := range warnings {
		_, err := tx.Exec(update, warning.WarningID)
		if err != nil {
			fmt.Printf("Error: %s, Update: %s, Params: %v\n", err.Error(), update, warning)
			tx.Rollback()
			return false
		}
	}

	err = tx.Commit()
	if err != nil {
		fmt.Printf("Error: %s, Update: %s\n", err.Error(), update)
		tx.Rollback()
		return false
	}

	return true
}

func (pgMembershipWarningRepository *PgMembershipWarningRepository) FindAll() []model.MembershipWarning {
	query :=
		"SELECT warning_id, user_id, username, first_warning_ts, grace_period_ns, is_relevant " +
			"FROM membership_warning"

	var warnings []model.MembershipWarning
	err := pgMembershipWarningRepository.db.Select(&warnings, query)
	if err != nil {
		fmt.Printf("Error: %s, Query: %s\n", err.Error(), query)
		return []model.MembershipWarning{}
	}

	return warnings
}

func (pgMembershipWarningRepository *PgMembershipWarningRepository) FindAllRelevant() []model.MembershipWarning {
	query :=
		"SELECT warning_id, user_id, username, first_warning_ts, grace_period_ns, is_relevant " +
			"FROM membership_warning " +
			"WHERE is_relevant = true"

	var warnings []model.MembershipWarning
	err := pgMembershipWarningRepository.db.Select(&warnings, query)
	if err != nil {
		fmt.Printf("Error: %s, Query: %s\n", err.Error(), query)
		return []model.MembershipWarning{}
	}

	return warnings
}
