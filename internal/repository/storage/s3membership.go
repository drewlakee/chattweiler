package storage

import (
	"bytes"
	"chattweiler/internal/logging"
	"chattweiler/internal/repository/model"
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/jszwec/csvutil"
	"io"
	"strings"
	"time"
)

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
		logging.Log.Error(logPackage, "CsvObjectStorageMembershipWarningRepository.getWarnings", err, "csv file reading error")
		return nil, err
	}

	var warnings []model.MembershipWarning
	err = csvutil.Unmarshal(csvFile, &warnings)
	if err != nil {
		logging.Log.Error(logPackage, "CsvObjectStorageMembershipWarningRepository.getWarnings", err, "csv file parsing error")
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
			logging.Log.Error(
				logPackage,
				"CsvObjectStorageMembershipWarningRepository.FindAllRelevant",
				err,
				"s3 client error: bucket - %s, key - %s", repo.bucket, previousKey,
			)
			return []model.MembershipWarning{}
		}

		allPreviousWarnings, err := repo.getWarnings(object.Body)
		if err != nil {
			logging.Log.Error(
				logPackage,
				"CsvObjectStorageMembershipWarningRepository.FindAllRelevant",
				err,
				"s3 client error: bucket - %s, key - %s", repo.bucket, previousKey,
			)
			return []model.MembershipWarning{}
		}

		relevantWarnings := repo.filterOnlyRelevant(allPreviousWarnings)
		if len(relevantWarnings) == 0 {
			return relevantWarnings
		}

		updatedCsvFile, err := csvutil.Marshal(relevantWarnings)
		if err != nil {
			logging.Log.Error(
				logPackage,
				"CsvObjectStorageMembershipWarningRepository.FindAllRelevant",
				err,
				"relevant warnings transformation to csv file error. bucket - %s, key - %s", repo.bucket, previousKey,
			)
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
			logging.Log.Error(
				logPackage,
				"CsvObjectStorageMembershipWarningRepository.FindAllRelevant",
				err,
				"csv file updating error. bucket - %s, key - %s", repo.bucket, newKey,
			)
			return []model.MembershipWarning{}
		}

		repo.currentDate = now
		logging.Log.Info(logPackage, "CsvObjectStorageMembershipWarningRepository.FindAllRelevant", "found for %d ms", time.Now().UnixMilli()-startTime)
		return relevantWarnings
	}

	currentKey := getDateAsString(repo.currentDate)
	object, err := repo.client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: &repo.bucket,
		Key:    &currentKey,
	})
	if err != nil {
		if !strings.Contains(err.Error(), "StatusCode: 404") {
			logging.Log.Error(
				logPackage,
				"CsvObjectStorageMembershipWarningRepository.FindAllRelevant",
				err,
				"s3 client error. bucket - %s, key - %s", repo.bucket, currentKey,
			)
		}
		return []model.MembershipWarning{}
	}

	warnings, err := repo.getWarnings(object.Body)
	if err != nil {
		logging.Log.Error(
			logPackage,
			"CsvObjectStorageMembershipWarningRepository.FindAllRelevant",
			err,
			"s3 client error. bucket - %s, key - %s", repo.bucket, currentKey,
		)
		return []model.MembershipWarning{}
	}

	relevantWarnings := repo.filterOnlyRelevant(warnings)
	logging.Log.Info(logPackage, "CsvObjectStorageMembershipWarningRepository.FindAllRelevant", "found for %d ms", time.Now().UnixMilli()-startTime)
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

		if err != nil && !strings.Contains(err.Error(), "NoSuchKey") {
			warningsToInsert, err = repo.getWarnings(object.Body)
			if err != nil {
				logging.Log.Error(
					logPackage,
					"CsvObjectStorageMembershipWarningRepository.Insert",
					err,
					"s3 client error: bucket - %s, key - %s", repo.bucket, currentKey,
				)
				return false
			}
		}
	}

	warningsToInsert = append(warningsToInsert, warning)
	currentKey := getDateAsString(repo.currentDate)
	updatedCsvFile, err := csvutil.Marshal(warningsToInsert)
	if err != nil {
		logging.Log.Error(
			logPackage,
			"CsvObjectStorageMembershipWarningRepository.Insert",
			err,
			"relevant warnings transformation to csv file error. bucket - %s, key - %s", repo.bucket, currentKey,
		)
		return false
	}

	newBody := bytes.NewReader(updatedCsvFile)
	_, err = repo.client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: &repo.bucket,
		Key:    &currentKey,
		Body:   newBody,
	})
	if err != nil {
		logging.Log.Error(
			logPackage,
			"CsvObjectStorageMembershipWarningRepository.Insert",
			err,
			"s3 client error. bucket - %s, key - %s", repo.bucket, currentKey,
		)
		return false
	}

	logging.Log.Info(logPackage, "CsvObjectStorageMembershipWarningRepository.Insert", "inserted for %d ms", time.Now().UnixMilli()-startTime)
	return true
}

func (repo *CsvObjectStorageMembershipWarningRepository) UpdateAllToIrrelevant(warnings ...model.MembershipWarning) bool {
	now := time.Now()
	startTime := now.UnixMilli()

	irrelevantWarnings := make(map[int]model.MembershipWarning)
	for _, warning := range warnings {
		irrelevantWarnings[warning.UserID] = warning
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
			logging.Log.Error(
				logPackage,
				"CsvObjectStorageMembershipWarningRepository.UpdateAllToIrrelevant",
				err,
				"s3 client error. bucket - %s, key - %s", repo.bucket, currentKey,
			)
			return false
		}

		warningsToUpdateArray, err = repo.getWarnings(object.Body)
		if err != nil {
			logging.Log.Error(
				logPackage,
				"CsvObjectStorageMembershipWarningRepository.UpdateAllToIrrelevant",
				err,
				"s3 client error. bucket - %s, key - %s", repo.bucket, currentKey,
			)
			return false
		}
	}

	for index, warning := range warningsToUpdateArray {
		if _, ok := irrelevantWarnings[warning.UserID]; ok {
			warningsToUpdateArray[index].IsRelevant = false
		}
	}

	currentKey := getDateAsString(repo.currentDate)
	updatedCsvFile, err := csvutil.Marshal(warningsToUpdateArray)
	if err != nil {
		logging.Log.Error(
			logPackage,
			"CsvObjectStorageMembershipWarningRepository.UpdateAllToIrrelevant",
			err,
			"relevant warnings transformation to csv file error. bucket - %s, key - %s", repo.bucket, currentKey,
		)
		return false
	}

	newBody := bytes.NewReader(updatedCsvFile)
	_, err = repo.client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: &repo.bucket,
		Key:    &currentKey,
		Body:   newBody,
	})
	if err != nil {
		logging.Log.Error(
			logPackage,
			"CsvObjectStorageMembershipWarningRepository.UpdateAllToIrrelevant",
			err,
			"csv file updating error",
		)
		return false
	}

	logging.Log.Info(logPackage, "CsvObjectStorageMembershipWarningRepository.UpdateAllToIrrelevant", "updated for %d ms", time.Now().UnixMilli()-startTime)
	return true
}

func getDateAsString(date time.Time) string {
	year, month, day := date.Date()
	return fmt.Sprintf("%d-%d-%d", year, day, month)
}
