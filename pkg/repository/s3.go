package repository

import (
	"bytes"
	"chattweiler/pkg/repository/model"
	"context"
	"fmt"
	"io"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/jszwec/csvutil"
	"github.com/sirupsen/logrus"
)

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
	phrases              []model.PhraseCsv
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

func (repo *CsvObjectStorageCachedPhraseRepository) castPhrases(phrases []model.PhraseCsv) []model.Phrase {
	result := make([]model.Phrase, len(phrases))
	for index, value := range phrases {
		result[index] = value
	}
	return result
}

func (repo *CsvObjectStorageCachedPhraseRepository) FindAll() []model.Phrase {
	startTime := time.Now().UnixMilli()
	if time.Now().Before(repo.lastCacheRefresh.Add(repo.cacheRefreshInterval)) {
		// atomic phrases read
		phrasesPtr := atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&repo.phrases)))
		if phrasesPtr != nil {
			phrases := *(*[]model.PhraseCsv)(phrasesPtr)
			if len(phrases) != 0 {
				return repo.castPhrases(phrases)
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

	var updatedPhrases []model.PhraseCsv
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
	return repo.castPhrases(updatedPhrases)
}

func (repo *CsvObjectStorageCachedPhraseRepository) FindAllByType(phraseType model.PhraseType) []model.Phrase {
	var phrases []model.Phrase
	for _, phrase := range repo.FindAll() {
		if phraseType == phrase.GetPhraseType() {
			phrases = append(phrases, phrase)
		}
	}
	return phrases
}

type CsvObjectStorageCachedContentCommandRepository struct {
	client               *s3.Client
	bucket               string
	key                  string
	cacheRefreshInterval time.Duration
	lastCacheRefresh     time.Time
	refreshMutex         sync.Mutex
	contentCommands      []model.ContentCommand
}

func NewCsvObjectStorageCachedContentSourceRepository(client *s3.Client, bucket, key string, cacheRefreshInterval time.Duration) *CsvObjectStorageCachedContentCommandRepository {
	return &CsvObjectStorageCachedContentCommandRepository{
		client:               client,
		bucket:               bucket,
		key:                  key,
		cacheRefreshInterval: cacheRefreshInterval,
		lastCacheRefresh:     time.Now(),
		contentCommands:      nil,
	}
}

func (repo *CsvObjectStorageCachedContentCommandRepository) FindAll() []model.ContentCommand {
	startTime := time.Now().UnixMilli()
	if time.Now().Before(repo.lastCacheRefresh.Add(repo.cacheRefreshInterval)) {
		// atomic content sources read
		contentSourcesPtr := atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&repo.contentCommands)))
		if contentSourcesPtr != nil {
			contentCommands := *(*[]model.ContentCommand)(contentSourcesPtr)
			if len(contentCommands) != 0 {
				return contentCommands
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
			"struct":   "CsvObjectStorageCachedContentCommandRepository",
			"func":     "FindAll",
			"err":      err,
			"bucket":   repo.bucket,
			"key":      repo.key,
			"fallback": "empty list",
		}).Error("s3 client error")
		return []model.ContentCommand{}
	}

	var updatedContentCommands []model.ContentCommand
	csvFile, err := io.ReadAll(object.Body)
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"struct":   "CsvObjectStorageCachedContentCommandRepository",
			"func":     "FindAll",
			"err":      err,
			"bucket":   repo.bucket,
			"key":      repo.key,
			"fallback": "empty list",
		}).Error("csv file reading error")
		return []model.ContentCommand{}
	}

	err = csvutil.Unmarshal(csvFile, &updatedContentCommands)
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"struct":   "CsvObjectStorageCachedContentCommandRepository",
			"func":     "FindAll",
			"err":      err,
			"bucket":   repo.bucket,
			"key":      repo.key,
			"fallback": "empty list",
		}).Error("csv file parsing error")
		return []model.ContentCommand{}
	}

	// atomic phrases write
	updatedPhrasesPtr := unsafe.Pointer(&updatedContentCommands)
	atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&repo.contentCommands)), updatedPhrasesPtr)

	repo.lastCacheRefresh = time.Now()
	logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
		"struct": "CsvObjectStorageCachedContentCommandRepository",
		"func":   "FindAll",
		"bucket": repo.bucket,
		"key":    repo.key,
	}).Info("Cache successfully updated for ", time.Now().UnixMilli()-startTime, "ms")
	return updatedContentCommands
}

func (repo *CsvObjectStorageCachedContentCommandRepository) FindByCommand(commandName string) *model.ContentCommand {
	for _, contentCommand := range repo.FindAll() {
		if strings.EqualFold(commandName, contentCommand.Name) {
			return &contentCommand
		}
	}
	return nil
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
	err = csvutil.Unmarshal(csvFile, &warnings)
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"struct": "CsvObjectStorageCachedContentCommandRepository",
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
				"struct":   "CsvObjectStorageCachedContentCommandRepository",
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
				"struct":   "CsvObjectStorageCachedContentCommandRepository",
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
			"struct": "CsvObjectStorageCachedContentCommandRepository",
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

func (repo *CsvObjectStorageMembershipWarningRepository) Insert(warning model.MembershipWarning) bool {
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

		if !strings.Contains(err.Error(), "NoSuchKey") {
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
	}

	warningsToInsert = append(warningsToInsert, warning)
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

	unRelevantWarnings := make(map[int]model.MembershipWarning)
	for _, warning := range warnings {
		unRelevantWarnings[warning.UserID] = warning
	}

	var warningsToUpdateArray []model.MembershipWarning
	if !isTheSameDate(repo.currentDate, now) {
		warningsToUpdateArray = repo.FindAllRelevant()
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

		warningsToUpdateArray, err = repo.getWarnings(object.Body)
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
	}

	for index, warning := range warningsToUpdateArray {
		if _, ok := unRelevantWarnings[warning.UserID]; ok {
			warningsToUpdateArray[index].IsRelevant = false
		}
	}

	if len(warningsToUpdateArray) == 0 {
		return true
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
