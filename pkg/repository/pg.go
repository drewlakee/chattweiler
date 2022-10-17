package repository

import (
	"chattweiler/pkg/logging"
	"chattweiler/pkg/repository/model"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/jmoiron/sqlx"
)

type CachedPgPhraseRepository struct {
	db                   *sqlx.DB
	cacheRefreshInterval time.Duration
	lastCacheRefresh     time.Time
	refreshMutex         sync.Mutex
	phrases              []model.PhrasePg
}

func NewCachedPgPhraseRepository(db *sqlx.DB, cacheRefreshInterval time.Duration) *CachedPgPhraseRepository {
	return &CachedPgPhraseRepository{
		db:                   db,
		cacheRefreshInterval: cacheRefreshInterval,
		lastCacheRefresh:     time.Now(),
		phrases:              nil,
	}
}

func (cachedPgPhraseRepository *CachedPgPhraseRepository) castPhrases(phrases []model.PhrasePg) []model.Phrase {
	result := make([]model.Phrase, len(phrases))
	for index, value := range phrases {
		result[index] = value
	}
	return result
}

func (cachedPgPhraseRepository *CachedPgPhraseRepository) FindAll() []model.Phrase {
	startTime := time.Now().UnixMilli()
	logging.Log.Info(logPackage, "CachedPgPhraseRepository.FindAll", "Updating phrases cache...")
	if time.Now().Before(cachedPgPhraseRepository.lastCacheRefresh.Add(cachedPgPhraseRepository.cacheRefreshInterval)) {
		// atomic phrases read
		phrasesPtr := atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&cachedPgPhraseRepository.phrases)))
		if phrasesPtr != nil {
			phrases := *(*[]model.PhrasePg)(phrasesPtr)
			if len(phrases) != 0 {
				return cachedPgPhraseRepository.castPhrases(phrases)
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

	var updatedPhrases []model.PhrasePg
	err := cachedPgPhraseRepository.db.Select(&updatedPhrases, query)
	if err != nil {
		logging.Log.Error(logPackage, "CachedPgPhraseRepository.FindAll", err, "Select error. query - %s", query)
		return []model.Phrase{}
	}

	// atomic phrases write
	updatedPhrasesPtr := unsafe.Pointer(&updatedPhrases)
	atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&cachedPgPhraseRepository.phrases)), updatedPhrasesPtr)

	cachedPgPhraseRepository.lastCacheRefresh = time.Now()
	logging.Log.Info(logPackage, "FindAll", fmt.Sprintf("Phrases cache successfully updated for %d ms", time.Now().UnixMilli()-startTime))
	return cachedPgPhraseRepository.castPhrases(updatedPhrases)
}

func (cachedPgPhraseRepository *CachedPgPhraseRepository) FindAllByType(phraseType model.PhraseType) []model.Phrase {
	var phrases []model.Phrase
	for _, phrase := range cachedPgPhraseRepository.FindAll() {
		if phraseType == phrase.GetPhraseType() {
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
		"INSERT INTO membership_warning (user_id, username, first_warning_ts, grace_period, is_relevant) " +
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
		logging.Log.Error(logPackage, "PgMembershipWarningRepository.Insert", err, "Insert error. query - %s, params - %v", insert, warning)
		return false
	}

	return true
}

func (pgMembershipWarningRepository *PgMembershipWarningRepository) UpdateAllToIrrelevant(warnings ...model.MembershipWarning) bool {
	update :=
		"UPDATE membership_warning " +
			"SET is_relevant=false " +
			"WHERE warning_id = $1"

	tx, err := pgMembershipWarningRepository.db.Begin()
	if err != nil {
		logging.Log.Error(logPackage, "PgMembershipWarningRepository.UpdateAllToIrrelevant", err, "Update error. query - %s", update)
		return false
	}

	for _, warning := range warnings {
		_, err := tx.Exec(update, warning.WarningID)
		if err != nil {
			logging.Log.Error(logPackage, "PgMembershipWarningRepository.UpdateAllToIrrelevant", err, "Update error. query - %s, params - %v", update, warning)
			tx.Rollback()
			return false
		}
	}

	err = tx.Commit()
	if err != nil {
		logging.Log.Error(logPackage, "PgMembershipWarningRepository.UpdateAllToIrrelevant", err, "Update error. query - %s", update)
		tx.Rollback()
		return false
	}

	return true
}

func (pgMembershipWarningRepository *PgMembershipWarningRepository) FindAllRelevant() []model.MembershipWarning {
	query :=
		"SELECT warning_id, user_id, username, first_warning_ts, grace_period, is_relevant " +
			"FROM membership_warning " +
			"WHERE is_relevant = true"

	var warnings []model.MembershipWarning
	err := pgMembershipWarningRepository.db.Select(&warnings, query)
	if err != nil {
		logging.Log.Error(logPackage, "PgMembershipWarningRepository.FindAllRelevant", err, "Select error. query - %s", query)
		return []model.MembershipWarning{}
	}

	return warnings
}

type CachedPgContentCommandRepository struct {
	db                   *sqlx.DB
	cacheRefreshInterval time.Duration
	lastCacheRefresh     time.Time
	refreshMutex         sync.Mutex
	contentSources       []model.ContentCommand
}

func NewCachedPgContentSourceRepository(db *sqlx.DB, cacheRefreshInterval time.Duration) *CachedPgContentCommandRepository {
	return &CachedPgContentCommandRepository{
		db:                   db,
		cacheRefreshInterval: cacheRefreshInterval,
		lastCacheRefresh:     time.Now(),
		contentSources:       nil,
	}
}

func (repo *CachedPgContentCommandRepository) FindAll() []model.ContentCommand {
	startTime := time.Now().UnixMilli()
	logging.Log.Info(logPackage, "CachedPgContentCommandRepository.FindAll", "Updating content commands cache...")
	if time.Now().Before(repo.lastCacheRefresh.Add(repo.cacheRefreshInterval)) {
		// atomic content sources read
		contentSourcesPtr := atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&repo.contentSources)))
		if contentSourcesPtr != nil {
			contentSources := *(*[]model.ContentCommand)(contentSourcesPtr)
			if len(contentSources) != 0 {
				return contentSources
			}
		}
	}

	// cache refresh lock
	repo.refreshMutex.Lock()
	defer repo.refreshMutex.Unlock()

	query :=
		"SELECT id, name, st.name AS media_type " +
			"FROM content_command AS cc, content_command_type AS cct " +
			"WHERE cc.media_type = cct.id "

	var updatedContentSources []model.ContentCommand
	err := repo.db.Select(&updatedContentSources, query)
	if err != nil {
		logging.Log.Info(logPackage, "CachedPgContentCommandRepository.FindAll", "Select error. query - %s", query)
		return []model.ContentCommand{}
	}

	// atomic content source write
	updatedUpdatedContentSourcesPtr := unsafe.Pointer(&updatedContentSources)
	atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&repo.contentSources)), updatedUpdatedContentSourcesPtr)

	repo.lastCacheRefresh = time.Now()
	logging.Log.Info(logPackage, "CachedPgContentCommandRepository.FindAll", "Content commands cache successfully updated for %d ms", time.Now().UnixMilli()-startTime)
	return updatedContentSources
}

func (repo *CachedPgContentCommandRepository) FindByCommandAlias(command string) *model.ContentCommand {
	for _, contentCommand := range repo.FindAll() {
		if contentCommand.ContainsAlias(command) {
			return &contentCommand
		}
	}
	return nil
}

func (repo *CachedPgContentCommandRepository) FindById(ID int) *model.ContentCommand {
	for _, contentCommand := range repo.FindAll() {
		if contentCommand.ID == ID {
			return &contentCommand
		}
	}
	return nil
}
