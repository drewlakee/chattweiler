package factory

import (
	"chattweiler/internal/configs"
	"chattweiler/internal/logging"
	"chattweiler/internal/repository"
	"chattweiler/internal/repository/storage"
	"chattweiler/internal/utils"
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var logPackage = "factory"

type StorageType string

const (
	CsvYandexObjectStorage StorageType = "csv_yandex_object_storage"
)

var objectStorageClientSingleton *s3.Client

func getObjectStorageClient() *s3.Client {
	if objectStorageClientSingleton != nil {
		return objectStorageClientSingleton
	}

	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			PartitionID:   "yc",
			URL:           "https://storage.yandexcloud.net",
			SigningRegion: region,
		}, nil
	})

	credentialsProvider := aws.CredentialsProviderFunc(func(ctx context.Context) (aws.Credentials, error) {
		return aws.Credentials{
			AccessKeyID:     utils.MustGetEnv(configs.YandexObjectStorageAccessKeyID),
			SecretAccessKey: utils.MustGetEnv(configs.YandexObjectStorageSecretAccessKey),
		}, nil
	})

	cfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithEndpointResolverWithOptions(customResolver),
		config.WithRegion(utils.MustGetEnv(configs.YandexObjectStorageRegion)),
		config.WithCredentialsProvider(credentialsProvider),
	)

	if err != nil {
		logging.Log.Panic(logPackage, "getObjectStorageClient", err, "storage config loading error")
	}

	objectStorageClientSingleton = s3.NewFromConfig(cfg)
	return objectStorageClientSingleton
}

func CreatePhraseRepository(repoType StorageType) repository.PhraseRepository {
	var repo repository.PhraseRepository
	switch repoType {
	case CsvYandexObjectStorage:
		fallthrough
	default:
		repo = createCsvObjectStorageCachedPhraseRepository()
	}

	return repo
}

func createCsvObjectStorageCachedPhraseRepository() *storage.CsvObjectStorageCachedPhraseRepository {
	cacheRefreshInterval, err := time.ParseDuration(utils.GetEnvOrDefault(configs.PhrasesCacheRefreshInterval))
	if err != nil {
		logging.Log.Panic(
			logPackage,
			"CsvObjectStorageCachedPhraseRepository.createCsvObjectStorageCachedPhraseRepository",
			err,
			configs.PhrasesCacheRefreshInterval.Key+": parsing of env variable is failed",
		)
	}

	return storage.NewCsvObjectStorageCachedPhraseRepository(
		getObjectStorageClient(),
		utils.MustGetEnv(configs.YandexObjectStoragePhrasesBucket),
		utils.MustGetEnv(configs.YandexObjectStoragePhrasesBucketKey),
		cacheRefreshInterval,
	)
}

func CreateContentSourceRepository(repoType StorageType) repository.CommandsRepository {
	var repo repository.CommandsRepository
	switch repoType {
	case CsvYandexObjectStorage:
		fallthrough
	default:
		repo = createCsvObjectStorageCachedContentSourceRepository()
	}

	return repo
}

func createCsvObjectStorageCachedContentSourceRepository() *storage.CsvObjectStorageCachedCommandRepository {
	cacheRefreshInterval, err := time.ParseDuration(utils.GetEnvOrDefault(configs.ContentCommandCacheRefreshInterval))
	if err != nil {
		logging.Log.Panic(
			logPackage,
			"CsvObjectStorageCachedCommandRepository.createCsvObjectStorageCachedContentSourceRepository",
			err,
			configs.ContentCommandCacheRefreshInterval.Key+": parsing of env variable is failed",
		)
	}

	return storage.NewCsvObjectStorageCachedCommandsRepository(
		getObjectStorageClient(),
		utils.MustGetEnv(configs.YandexObjectStorageContentSourceBucket),
		utils.MustGetEnv(configs.YandexObjectStorageContentSourceBucketKey),
		cacheRefreshInterval,
	)
}

func CreateMembershipWarningRepository(repoType StorageType) repository.MembershipWarningRepository {
	var repo repository.MembershipWarningRepository
	switch repoType {
	case CsvYandexObjectStorage:
		fallthrough
	default:
		repo = createCsvObjectStorageMembershipWarningRepository()
	}

	return repo
}

func createCsvObjectStorageMembershipWarningRepository() *storage.CsvObjectStorageMembershipWarningRepository {
	return storage.NewCsvObjectStorageMembershipWarningRepository(
		getObjectStorageClient(),
		utils.MustGetEnv(configs.YandexObjectStorageMembershipWarningBucket),
	)
}
