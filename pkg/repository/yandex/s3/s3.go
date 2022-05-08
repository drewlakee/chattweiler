package s3

import (
	"bytes"
	"chattweiler/pkg/repository/model"
	"chattweiler/pkg/repository/model/types"
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/jszwec/csvutil"
	"github.com/sirupsen/logrus"
	"io"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

var packageLogFields = logrus.Fields{
	"package": "s3",
}

func getDateAsString(date time.Time) string {
	year, month, day := date.Date()
	return fmt.Sprintf("%d-%d-%d", year, day, month)
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

type CsvObjectStorageCachedPhraseRepository struct {
	client               *s3.Client
	bucket               string
	key                  string
	cacheRefreshInterval time.Duration
	lastCacheRefresh     time.Time
	refreshMutex         sync.Mutex
	phrases              []model.Phrase
}

func NewCsvObjectStorageCachedPhraseRepository(client *s3.Client, bucket, key string, cacheRefreshInterval time.Duration) *CsvObjectStorageCachedPhraseRepository {
	return &CsvObjectStorageCachedPhraseRepository{
		client:               client,
		bucket:               bucket,
		key:                  key,
		cacheRefreshInterval: cacheRefreshInterval,
		lastCacheRefresh:     time.Now(),
		phrases:              nil,
	}
}

func (repo *CsvObjectStorageCachedPhraseRepository) FindAll() []model.Phrase {
	startTime := time.Now().UnixMilli()
	if time.Now().Before(repo.lastCacheRefresh.Add(repo.cacheRefreshInterval)) {
		// atomic phrases read
		phrasesPtr := atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&repo.phrases)))
		if phrasesPtr != nil {
			phrases := *(*[]model.Phrase)(phrasesPtr)
			if len(phrases) != 0 {
				return phrases
			}
		}
	}

	// cache refresh lock
	repo.refreshMutex.Lock()
	defer repo.refreshMutex.Unlock()

	object, err := repo.client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: &repo.bucket,
		Key:    &repo.key,
	})
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"struct":   "CsvObjectStorageCachedPhraseRepository",
			"func":     "FindAll",
			"err":      err,
			"bucket":   repo.bucket,
			"key":      repo.key,
			"fallback": "empty list",
		}).Error("s3 client error")
		return []model.Phrase{}
	}

	var updatedPhrases []model.Phrase
	csvFile, err := io.ReadAll(object.Body)
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"struct":   "CsvObjectStorageCachedPhraseRepository",
			"func":     "FindAll",
			"err":      err,
			"bucket":   repo.bucket,
			"key":      repo.key,
			"fallback": "empty list",
		}).Error("csv file reading error")
		return []model.Phrase{}
	}

	err = csvutil.Unmarshal(csvFile, &updatedPhrases)
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"struct":   "CsvObjectStorageCachedPhraseRepository",
			"func":     "FindAll",
			"err":      err,
			"bucket":   repo.bucket,
			"key":      repo.key,
			"fallback": "empty list",
		}).Error("csv file parsing error")
		return []model.Phrase{}
	}

	// atomic phrases write
	updatedPhrasesPtr := unsafe.Pointer(&updatedPhrases)
	atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&repo.phrases)), updatedPhrasesPtr)

	repo.lastCacheRefresh = time.Now()
	logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
		"struct": "CsvObjectStorageCachedPhraseRepository",
		"func":   "FindAll",
		"bucket": repo.bucket,
		"key":    repo.key,
	}).Info("Cache successfully updated for ", time.Now().UnixMilli()-startTime, "ms")
	return updatedPhrases
}

func (repo *CsvObjectStorageCachedPhraseRepository) FindAllByType(phraseType types.PhraseType) []model.Phrase {
	var phrases []model.Phrase
	for _, phrase := range repo.FindAll() {
		if phraseType == phrase.PhraseType {
			phrases = append(phrases, phrase)
		}
	}
	return phrases
}

type CsvObjectStorageCachedContentSourceRepository struct {
	client               *s3.Client
	bucket               string
	key                  string
	cacheRefreshInterval time.Duration
	lastCacheRefresh     time.Time
	refreshMutex         sync.Mutex
	contentSources       []model.ContentSource
}

func NewCsvObjectStorageCachedContentSourceRepository(client *s3.Client, bucket, key string, cacheRefreshInterval time.Duration) *CsvObjectStorageCachedContentSourceRepository {
	return &CsvObjectStorageCachedContentSourceRepository{
		client:               client,
		bucket:               bucket,
		key:                  key,
		cacheRefreshInterval: cacheRefreshInterval,
		lastCacheRefresh:     time.Now(),
		contentSources:       nil,
	}
}

func (repo *CsvObjectStorageCachedContentSourceRepository) FindAll() []model.ContentSource {
	startTime := time.Now().UnixMilli()
	if time.Now().Before(repo.lastCacheRefresh.Add(repo.cacheRefreshInterval)) {
		// atomic content sources read
		contentSourcesPtr := atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&repo.contentSources)))
		if contentSourcesPtr != nil {
			contentSources := *(*[]model.ContentSource)(contentSourcesPtr)
			if len(contentSources) != 0 {
				return contentSources
			}
		}
	}

	// cache refresh lock
	repo.refreshMutex.Lock()
	defer repo.refreshMutex.Unlock()

	object, err := repo.client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: &repo.bucket,
		Key:    &repo.key,
	})
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"struct":   "CsvObjectStorageCachedContentSourceRepository",
			"func":     "FindAll",
			"err":      err,
			"bucket":   repo.bucket,
			"key":      repo.key,
			"fallback": "empty list",
		}).Error("s3 client error")
		return []model.ContentSource{}
	}

	var updatedContentSources []model.ContentSource
	csvFile, err := io.ReadAll(object.Body)
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"struct":   "CsvObjectStorageCachedContentSourceRepository",
			"func":     "FindAll",
			"err":      err,
			"bucket":   repo.bucket,
			"key":      repo.key,
			"fallback": "empty list",
		}).Error("csv file reading error")
		return []model.ContentSource{}
	}

	err = csvutil.Unmarshal(csvFile, &updatedContentSources)
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"struct":   "CsvObjectStorageCachedContentSourceRepository",
			"func":     "FindAll",
			"err":      err,
			"bucket":   repo.bucket,
			"key":      repo.key,
			"fallback": "empty list",
		}).Error("csv file parsing error")
		return []model.ContentSource{}
	}

	// atomic phrases write
	updatedPhrasesPtr := unsafe.Pointer(&updatedContentSources)
	atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&repo.contentSources)), updatedPhrasesPtr)

	repo.lastCacheRefresh = time.Now()
	logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
		"struct": "CsvObjectStorageCachedContentSourceRepository",
		"func":   "FindAll",
		"bucket": repo.bucket,
		"key":    repo.key,
	}).Info("Cache successfully updated for ", time.Now().UnixMilli()-startTime, "ms")
	return updatedContentSources
}

func (repo *CsvObjectStorageCachedContentSourceRepository) FindAllByType(sourceType types.ContentSourceType) []model.ContentSource {
	var contentSources []model.ContentSource
	for _, contentSource := range repo.FindAll() {
		if sourceType == contentSource.SourceType {
			contentSources = append(contentSources, contentSource)
		}
	}
	return contentSources
}

type CsvObjectStorageMembershipWarningRepository struct {
	client      *s3.Client
	bucket      string
	currentDate time.Time
}

func NewCsvObjectStorageMembershipWarningRepository(client *s3.Client, bucket string) *CsvObjectStorageMembershipWarningRepository {
	return &CsvObjectStorageMembershipWarningRepository{
		client:      client,
		bucket:      bucket,
		currentDate: time.Now(),
	}
}

func (repo *CsvObjectStorageMembershipWarningRepository) getWarnings(csvFileReader io.ReadCloser) ([]model.MembershipWarning, error) {
	csvFile, err := io.ReadAll(csvFileReader)
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"struct": "CsvObjectStorageMembershipWarningRepository",
			"func":   "getWarnings",
		}).Error("csv file reading error")
		return nil, err
	}

	var warnings []model.MembershipWarning
	err = csvutil.Unmarshal(csvFile, warnings)
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"struct": "CsvObjectStorageCachedContentSourceRepository",
			"func":   "getWarnings",
		}).Error("csv file parsing error")
		return nil, err
	}

	return warnings, nil
}

func (repo *CsvObjectStorageMembershipWarningRepository) filterOnlyRelevant(warnings []model.MembershipWarning) []model.MembershipWarning {
	var relevantWarnings []model.MembershipWarning
	for _, warning := range warnings {
		if warning.IsRelevant {
			relevantWarnings = append(relevantWarnings, warning)
		}
	}
	return relevantWarnings
}

func (repo *CsvObjectStorageMembershipWarningRepository) FindAllRelevant() []model.MembershipWarning {
	now := time.Now()
	startTime := now.UnixMilli()
	if !isTheSameDate(repo.currentDate, now) {
		previousKey := getDateAsString(repo.currentDate)
		object, err := repo.client.GetObject(context.TODO(), &s3.GetObjectInput{
			Bucket: &repo.bucket,
			Key:    &previousKey,
		})

		if err != nil {
			logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
				"struct":   "CsvObjectStorageMembershipWarningRepository",
				"func":     "FindAllRelevant",
				"err":      err,
				"bucket":   repo.bucket,
				"key":      previousKey,
				"fallback": "empty list",
			}).Error("s3 client error")
			return []model.MembershipWarning{}
		}

		allPreviousWarnings, err := repo.getWarnings(object.Body)
		if err != nil {
			logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
				"struct":   "CsvObjectStorageMembershipWarningRepository",
				"func":     "FindAllRelevant",
				"err":      err,
				"bucket":   repo.bucket,
				"key":      previousKey,
				"fallback": "empty list",
			}).Error("s3 client error")
			return []model.MembershipWarning{}
		}

		relevantWarnings := repo.filterOnlyRelevant(allPreviousWarnings)
		updatedCsvFile, err := csvutil.Marshal(relevantWarnings)
		if err != nil {
			logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
				"struct":   "CsvObjectStorageCachedContentSourceRepository",
				"func":     "FindAllRelevant",
				"err":      err,
				"bucket":   repo.bucket,
				"key":      previousKey,
				"fallback": "empty list",
			}).Error("relevant warnings transformation to csv file error")
			return []model.MembershipWarning{}
		}

		newKey := getDateAsString(now)
		newBody := bytes.NewReader(updatedCsvFile)
		_, err = repo.client.PutObject(context.TODO(), &s3.PutObjectInput{
			Bucket: &repo.bucket,
			Key:    &newKey,
			Body:   newBody,
		})
		if err != nil {
			logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
				"struct":   "CsvObjectStorageCachedContentSourceRepository",
				"func":     "FindAllRelevant",
				"err":      err,
				"bucket":   repo.bucket,
				"key":      newKey,
				"fallback": "empty list",
			}).Error("csv file updating error")
			return []model.MembershipWarning{}
		}

		repo.currentDate = now
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"struct": "CsvObjectStorageCachedContentSourceRepository",
			"func":   "FindAllRelevant",
			"bucket": repo.bucket,
			"key":    newKey,
		}).Info("Found for ", time.Now().UnixMilli()-startTime, "ms")
		return relevantWarnings
	}

	currentKey := getDateAsString(repo.currentDate)
	object, err := repo.client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: &repo.bucket,
		Key:    &currentKey,
	})
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"struct":   "CsvObjectStorageMembershipWarningRepository",
			"func":     "FindAllRelevant",
			"err":      err,
			"bucket":   repo.bucket,
			"key":      currentKey,
			"fallback": "empty list",
		}).Error("s3 client error")
		return []model.MembershipWarning{}
	}

	warnings, err := repo.getWarnings(object.Body)
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"struct":   "CsvObjectStorageMembershipWarningRepository",
			"func":     "FindAllRelevant",
			"err":      err,
			"bucket":   repo.bucket,
			"key":      currentKey,
			"fallback": "empty list",
		}).Error("s3 client error")
		return []model.MembershipWarning{}
	}

	relevantWarnings := repo.filterOnlyRelevant(warnings)
	logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
		"struct": "CsvObjectStorageMembershipWarningRepository",
		"func":   "FindAllRelevant",
		"bucket": repo.bucket,
		"key":    currentKey,
	}).Info("Found for ", time.Now().UnixMilli()-startTime, "ms")
	return relevantWarnings
}

func (repo *CsvObjectStorageMembershipWarningRepository) Insert(model.MembershipWarning) bool {
	now := time.Now()
	startTime := now.UnixMilli()
	var warningsToInsert []model.MembershipWarning
	if !isTheSameDate(repo.currentDate, now) {
		warningsToInsert = repo.FindAllRelevant()
	} else {
		currentKey := getDateAsString(repo.currentDate)
		object, err := repo.client.GetObject(context.TODO(), &s3.GetObjectInput{
			Bucket: &repo.bucket,
			Key:    &currentKey,
		})
		if err != nil {
			logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
				"struct": "CsvObjectStorageMembershipWarningRepository",
				"func":   "Insert",
				"err":    err,
				"bucket": repo.bucket,
				"key":    currentKey,
			}).Error("s3 client error")
			return false
		}

		warningsToInsert, err = repo.getWarnings(object.Body)
		if err != nil {
			logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
				"struct": "CsvObjectStorageMembershipWarningRepository",
				"func":   "Insert",
				"err":    err,
				"bucket": repo.bucket,
				"key":    currentKey,
			}).Error("s3 client error")
			return false
		}
	}

	currentKey := getDateAsString(repo.currentDate)
	updatedCsvFile, err := csvutil.Marshal(warningsToInsert)
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"struct": "CsvObjectStorageMembershipWarningRepository",
			"func":   "Insert",
			"err":    err,
			"bucket": repo.bucket,
			"key":    currentKey,
		}).Error("relevant warnings transformation to csv file error")
		return false
	}

	newBody := bytes.NewReader(updatedCsvFile)
	_, err = repo.client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: &repo.bucket,
		Key:    &currentKey,
		Body:   newBody,
	})
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"struct": "CsvObjectStorageMembershipWarningRepository",
			"func":   "Insert",
			"err":    err,
			"bucket": repo.bucket,
			"key":    currentKey,
		}).Error("csv file updating error")
		return false
	}

	logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
		"struct": "CsvObjectStorageMembershipWarningRepository",
		"func":   "Insert",
		"bucket": repo.bucket,
		"key":    currentKey,
	}).Info("Inserted for ", time.Now().UnixMilli()-startTime, "ms")
	return true
}

func (repo *CsvObjectStorageMembershipWarningRepository) UpdateAllToUnRelevant(warnings ...model.MembershipWarning) bool {
	now := time.Now()
	startTime := now.UnixMilli()
	var warningsToUpdateMap map[int]model.MembershipWarning
	if !isTheSameDate(repo.currentDate, now) {
		for _, warning := range repo.FindAllRelevant() {
			warningsToUpdateMap[warning.UserID] = warning
		}
	} else {
		currentKey := getDateAsString(repo.currentDate)
		object, err := repo.client.GetObject(context.TODO(), &s3.GetObjectInput{
			Bucket: &repo.bucket,
			Key:    &currentKey,
		})
		if err != nil {
			logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
				"struct": "CsvObjectStorageMembershipWarningRepository",
				"func":   "UpdateAllToUnRelevant",
				"err":    err,
				"bucket": repo.bucket,
				"key":    currentKey,
			}).Error("s3 client error")
			return false
		}

		warnings, err = repo.getWarnings(object.Body)
		if err != nil {
			logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
				"struct": "CsvObjectStorageMembershipWarningRepository",
				"func":   "UpdateAllToUnRelevant",
				"err":    err,
				"bucket": repo.bucket,
				"key":    currentKey,
			}).Error("s3 client error")
			return false
		}

		for _, warning := range warnings {
			warningsToUpdateMap[warning.UserID] = warning
		}
	}

	for _, warning := range warnings {
		if existingWarning, ok := warningsToUpdateMap[warning.UserID]; ok {
			existingWarning.IsRelevant = false
		}
	}

	warningsToUpdateArray := make([]model.MembershipWarning, len(warningsToUpdateMap))
	index := 0
	for _, warning := range warningsToUpdateMap {
		warningsToUpdateArray[index] = warning
		index++
	}

	currentKey := getDateAsString(repo.currentDate)
	updatedCsvFile, err := csvutil.Marshal(warningsToUpdateArray)
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"struct": "CsvObjectStorageMembershipWarningRepository",
			"func":   "UpdateAllToUnRelevant",
			"err":    err,
			"bucket": repo.bucket,
			"key":    currentKey,
		}).Error("relevant warnings transformation to csv file error")
		return false
	}

	newBody := bytes.NewReader(updatedCsvFile)
	_, err = repo.client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: &repo.bucket,
		Key:    &currentKey,
		Body:   newBody,
	})
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"struct": "CsvObjectStorageMembershipWarningRepository",
			"func":   "UpdateAllToUnRelevant",
			"err":    err,
			"bucket": repo.bucket,
			"key":    currentKey,
		}).Error("csv file updating error")
		return false
	}

	logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
		"struct": "CsvObjectStorageMembershipWarningRepository",
		"func":   "UpdateAllToUnRelevant",
		"bucket": repo.bucket,
		"key":    currentKey,
	}).Info("Updated for ", time.Now().UnixMilli()-startTime, "ms")
	return true
}
