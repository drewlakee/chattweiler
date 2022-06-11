package repository

import (
	"chattweiler/pkg/configs"
	"chattweiler/pkg/utils"
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

type RepositoryType string

const (
	Postgresql             RepositoryType = "postgresql"
	CsvYandexObjectStorage RepositoryType = "csv_yandex_object_storage"
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
		logrus.WithFields(packageLogFields).WithFields(packageLogFields).WithFields(logrus.Fields{
			"func": "getPostgresqlConnection",
			"err":  err,
		}).Fatal("postgres connection error")
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
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"func": "getObjectStorageClient",
			"err":  err,
		}).Fatal("storage config loading error")
	}

	objectStorageClientSingleton = s3.NewFromConfig(cfg)
	return objectStorageClientSingleton
}

func CreatePhraseRepository(repoType RepositoryType) PhraseRepository {
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
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"func": "createCsvObjectStorageCachedPhraseRepository",
			"err":  err,
			"key":  configs.PhrasesCacheRefreshInterval.Key,
		}).Fatal("parsing of env variable is failed")
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
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"func": "createPostgresqlCachedPhraseRepository",
			"err":  err,
			"key":  configs.PhrasesCacheRefreshInterval.Key,
		}).Fatal("parsing of env variable is failed")
	}

	return NewCachedPgPhraseRepository(getPostgresqlConnection(), pgPhrasesCacheRefreshInterval)
}

func CreateContentSourceRepository(repoType RepositoryType) ContentSourceRepository {
	var repo ContentSourceRepository
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

func createCsvObjectStorageCachedContentSourceRepository() *CsvObjectStorageCachedContentSourceRepository {
	cacheRefreshInterval, err := time.ParseDuration(utils.GetEnvOrDefault(configs.ContentSourceCacheRefreshInterval))
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"func": "createCsvObjectStorageCachedContentSourceRepository",
			"err":  err,
			"key":  configs.ContentSourceCacheRefreshInterval.Key,
		}).Fatal("parsing of env variable is failed")
	}

	return NewCsvObjectStorageCachedContentSourceRepository(
		getObjectStorageClient(),
		utils.MustGetEnv(configs.YandexObjectStorageContentSourceBucket),
		utils.MustGetEnv(configs.YandexObjectStorageContentSourceBucketKey),
		cacheRefreshInterval,
	)
}

func createPostgresqlCachedContentSourceRepository() ContentSourceRepository {
	pgContentSourceCacheRefreshInterval, err := time.ParseDuration(utils.GetEnvOrDefault(configs.ContentSourceCacheRefreshInterval))
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"func": "createPostgresqlCachedContentSourceRepository",
			"err":  err,
			"key":  configs.ContentSourceCacheRefreshInterval.Key,
		}).Fatal("parsing of env variable is failed")
	}

	return NewCachedPgContentSourceRepository(getPostgresqlConnection(), pgContentSourceCacheRefreshInterval)
}

func CreateMembershipWarningRepository(repoType RepositoryType) MembershipWarningRepository {
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
