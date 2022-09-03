package repository

import (
	"chattweiler/pkg/configs"
	"chattweiler/pkg/logging"
	"chattweiler/pkg/utils"
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/jmoiron/sqlx"
)

type StorageType string

const (
	Postgresql             StorageType = "postgresql"
	CsvYandexObjectStorage StorageType = "csv_yandex_object_storage"
)

var pgConnectionSingleton *sqlx.DB
var objectStorageClientSingleton *s3.Client

func getPostgresqlConnection() *sqlx.DB {
	pgDataSourceString := utils.MustGetEnv(configs.PgDatasourceString)

	if pgConnectionSingleton != nil {
		return pgConnectionSingleton
	}

	pgConnectionSingleton, err := sqlx.Connect("postgres", pgDataSourceString)
	if err != nil {
		logging.Log.Panic(logPackage, "getPostgresqlConnection", err, "postgres connection error")
	}

	return pgConnectionSingleton
}

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

func CreatePhraseRepository(repoType StorageType) PhraseRepository {
	var repo PhraseRepository
	switch repoType {
	case CsvYandexObjectStorage:
		repo = createCsvObjectStorageCachedPhraseRepository()
	case Postgresql:
		fallthrough
	default:
		repo = createPostgresqlCachedPhraseRepository()
	}

	return repo
}

func createCsvObjectStorageCachedPhraseRepository() *CsvObjectStorageCachedPhraseRepository {
	cacheRefreshInterval, err := time.ParseDuration(utils.GetEnvOrDefault(configs.PhrasesCacheRefreshInterval))
	if err != nil {
		logging.Log.Panic(
			logPackage,
			"CsvObjectStorageCachedPhraseRepository.createCsvObjectStorageCachedPhraseRepository",
			err,
			configs.PhrasesCacheRefreshInterval.Key+": parsing of env variable is failed",
		)
	}

	return NewCsvObjectStorageCachedPhraseRepository(
		getObjectStorageClient(),
		utils.MustGetEnv(configs.YandexObjectStoragePhrasesBucket),
		utils.MustGetEnv(configs.YandexObjectStoragePhrasesBucketKey),
		cacheRefreshInterval,
	)
}

func createPostgresqlCachedPhraseRepository() PhraseRepository {
	pgPhrasesCacheRefreshInterval, err := time.ParseDuration(utils.GetEnvOrDefault(configs.PhrasesCacheRefreshInterval))
	if err != nil {
		logging.Log.Panic(
			logPackage,
			"PhraseRepository.createPostgresqlCachedPhraseRepository",
			err,
			configs.PhrasesCacheRefreshInterval.Key+": parsing of env variable is failed",
		)
	}

	return NewCachedPgPhraseRepository(getPostgresqlConnection(), pgPhrasesCacheRefreshInterval)
}

func CreateContentSourceRepository(repoType StorageType) ContentCommandRepository {
	var repo ContentCommandRepository
	switch repoType {
	case CsvYandexObjectStorage:
		repo = createCsvObjectStorageCachedContentSourceRepository()
	case Postgresql:
		fallthrough
	default:
		repo = createPostgresqlCachedContentSourceRepository()
	}

	return repo
}

func createCsvObjectStorageCachedContentSourceRepository() *CsvObjectStorageCachedContentCommandRepository {
	cacheRefreshInterval, err := time.ParseDuration(utils.GetEnvOrDefault(configs.ContentCommandCacheRefreshInterval))
	if err != nil {
		logging.Log.Panic(
			logPackage,
			"CsvObjectStorageCachedContentCommandRepository.createCsvObjectStorageCachedContentSourceRepository",
			err,
			configs.ContentCommandCacheRefreshInterval.Key+": parsing of env variable is failed",
		)
	}

	return NewCsvObjectStorageCachedContentSourceRepository(
		getObjectStorageClient(),
		utils.MustGetEnv(configs.YandexObjectStorageContentSourceBucket),
		utils.MustGetEnv(configs.YandexObjectStorageContentSourceBucketKey),
		cacheRefreshInterval,
	)
}

func createPostgresqlCachedContentSourceRepository() ContentCommandRepository {
	pgContentSourceCacheRefreshInterval, err := time.ParseDuration(utils.GetEnvOrDefault(configs.ContentCommandCacheRefreshInterval))
	if err != nil {
		logging.Log.Panic(
			logPackage,
			"ContentCommandRepository.createPostgresqlCachedContentSourceRepository",
			err,
			configs.ContentCommandCacheRefreshInterval.Key+": parsing of env variable is failed",
		)
	}

	return NewCachedPgContentSourceRepository(getPostgresqlConnection(), pgContentSourceCacheRefreshInterval)
}

func CreateMembershipWarningRepository(repoType StorageType) MembershipWarningRepository {
	var repo MembershipWarningRepository
	switch repoType {
	case CsvYandexObjectStorage:
		repo = createCsvObjectStorageMembershipWarningRepository()
	case Postgresql:
		fallthrough
	default:
		repo = NewPgMembershipWarningRepository(getPostgresqlConnection())
	}

	return repo
}

func createCsvObjectStorageMembershipWarningRepository() *CsvObjectStorageMembershipWarningRepository {
	return NewCsvObjectStorageMembershipWarningRepository(
		getObjectStorageClient(),
		utils.MustGetEnv(configs.YandexObjectStorageMembershipWarningBucket),
	)
}
