package storage

import (
	"chattweiler/internal/logging"
	"chattweiler/internal/repository/model"
	"context"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/jszwec/csvutil"
	"io"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

type CsvObjectStorageCachedPhraseRepository struct {
	client               *s3.Client
	bucket               string
	key                  string
	cacheRefreshInterval time.Duration
	lastCacheRefresh     time.Time
	refreshMutex         sync.Mutex

	cachedList        []model.Phrase
	cachedListsByType map[model.PhraseType][]model.Phrase
}

func NewCsvObjectStorageCachedPhraseRepository(client *s3.Client, bucket, key string, cacheRefreshInterval time.Duration) *CsvObjectStorageCachedPhraseRepository {
	repository := CsvObjectStorageCachedPhraseRepository{
		client:               client,
		bucket:               bucket,
		key:                  key,
		cacheRefreshInterval: cacheRefreshInterval,
		lastCacheRefresh:     time.Now(),
	}
	err := repository.refreshCache()
	if err != nil {
		panic(err)
	}
	return &repository
}

func (repo *CsvObjectStorageCachedPhraseRepository) isNeededInvalidateCache() bool {
	return time.Now().After(repo.lastCacheRefresh.Add(repo.cacheRefreshInterval))
}

func (repo *CsvObjectStorageCachedPhraseRepository) refreshCache() error {
	startTime := time.Now().UnixMilli()

	// cache refresh lock
	repo.refreshMutex.Lock()
	defer repo.refreshMutex.Unlock()

	object, err := repo.client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: &repo.bucket,
		Key:    &repo.key,
	})
	if err != nil {
		logging.Log.Error(
			logPackage,
			"CsvObjectStorageCachedPhraseRepository.refreshCache",
			err,
			"s3 client error: bucket - %s, key - %s", repo.bucket, repo.key,
		)
		return err
	}

	csvFile, err := io.ReadAll(object.Body)
	if err != nil {
		logging.Log.Error(logPackage, "CsvObjectStorageCachedPhraseRepository.refreshCache", err, "csv file reading error")
		return err
	}

	var csvPhrases []model.PhraseCsv
	err = csvutil.Unmarshal(csvFile, &csvPhrases)
	if err != nil {
		logging.Log.Error(logPackage, "CsvObjectStorageCachedPhraseRepository.refreshCache", err, "csv file parsing error")
		return err
	}

	var list = convertCsvPhrases(csvPhrases)

	var mapByType = make(map[model.PhraseType][]model.Phrase)
	for _, phrase := range list {
		mapByType[phrase.GetPhraseType()] = append(mapByType[phrase.GetPhraseType()], phrase)
	}

	listPtr := unsafe.Pointer(&list)
	atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&repo.cachedList)), listPtr)

	mapByTypePtr := unsafe.Pointer(&mapByType)
	atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&repo.cachedListsByType)), mapByTypePtr)

	repo.lastCacheRefresh = time.Now()
	logging.Log.Info(logPackage, "CsvObjectStorageCachedPhraseRepository.refreshCache", "Cache successfully updated for %d ms", time.Now().UnixMilli()-startTime)
	return nil
}

func (repo *CsvObjectStorageCachedPhraseRepository) FindAll() []model.Phrase {
	if repo.isNeededInvalidateCache() {
		err := repo.refreshCache()
		if err != nil {
			return []model.Phrase{}
		}
	}

	ptr := atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&repo.cachedList)))
	if ptr != nil {
		phrases := *(*[]model.Phrase)(ptr)
		if len(phrases) != 0 {
			return phrases
		}
	}
	return nil
}

func (repo *CsvObjectStorageCachedPhraseRepository) FindAllByType(phraseType model.PhraseType) []model.Phrase {
	if repo.isNeededInvalidateCache() {
		err := repo.refreshCache()
		if err != nil {
			return []model.Phrase{}
		}
	}

	ptr := atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&repo.cachedListsByType)))
	if ptr != nil {
		phrases := *(*map[model.PhraseType][]model.Phrase)(ptr)
		if phrasesByType, exists := phrases[phraseType]; exists {
			return phrasesByType
		}
	}
	return []model.Phrase{}
}

func isTheSameDate(first, second time.Time) bool {
	year1, month1, day1 := first.Date()
	year2, month2, day2 := second.Date()

	if year1 != year2 {
		return false
	} else if month1 != month2 {
		return false
	} else if day1 != day2 {
		return false
	}

	return true
}

func convertCsvPhrases(phrases []model.PhraseCsv) []model.Phrase {
	result := make([]model.Phrase, len(phrases))
	for index, value := range phrases {
		result[index] = value
	}
	return result
}
