package factory

import (
	"chattweiler/pkg/app/configs/static"
	"chattweiler/pkg/app/utils"
	"chattweiler/pkg/repository"
	"chattweiler/pkg/repository/pg"
	objectstorage "chattweiler/pkg/repository/yandex/s3"
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

var packageLogFields = logrus.Fields{
	"package": "factory",
}

type RepositoryType string

const (
	Postgresql             RepositoryType = "postgresql"
	CsvYandexObjectStorage RepositoryType = "csv_yandex_object_storage"
)

var pgConnectionSingleton *sqlx.DB
var objectStorageClientSingleton *s3.Client

func getPostgresqlConnection() *sqlx.DB {
	pgDataSourceString := utils.MustGetEnv(static.PgDatasourceString)

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
			AccessKeyID:     utils.MustGetEnv(static.YandexObjectStorageAccessKeyID),
			SecretAccessKey: utils.MustGetEnv(static.YandexObjectStorageSecretAccessKey),
		}, nil
	})

	cfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithEndpointResolverWithOptions(customResolver),
		config.WithRegion(utils.MustGetEnv(static.YandexObjectStorageRegion)),
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

func CreatePhraseRepository(repoType RepositoryType) repository.PhraseRepository {
	var repo repository.PhraseRepository
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

func createCsvObjectStorageCachedPhraseRepository() *objectstorage.CsvObjectStorageCachedPhraseRepository {
	cacheRefreshInterval, err := time.ParseDuration(utils.GetEnvOrDefault(static.PhrasesCacheRefreshInterval))
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"func": "createCsvObjectStorageCachedPhraseRepository",
			"err":  err,
			"key":  static.PhrasesCacheRefreshInterval.Key,
		}).Fatal("parsing of env variable is failed")
	}

	return objectstorage.NewCsvObjectStorageCachedPhraseRepository(
		getObjectStorageClient(),
		utils.MustGetEnv(static.YandexObjectStoragePhrasesBucket),
		utils.MustGetEnv(static.YandexObjectStoragePhrasesBucketKey),
		cacheRefreshInterval,
	)
}

func createPostgresqlCachedPhraseRepository() repository.PhraseRepository {
	pgPhrasesCacheRefreshInterval, err := time.ParseDuration(utils.GetEnvOrDefault(static.PhrasesCacheRefreshInterval))
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"func": "createPostgresqlCachedPhraseRepository",
			"err":  err,
			"key":  static.PhrasesCacheRefreshInterval.Key,
		}).Fatal("parsing of env variable is failed")
	}

	return pg.NewCachedPgPhraseRepository(getPostgresqlConnection(), pgPhrasesCacheRefreshInterval)
}

func CreateContentSourceRepository(repoType RepositoryType) repository.ContentSourceRepository {
	var repo repository.ContentSourceRepository
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

func createCsvObjectStorageCachedContentSourceRepository() *objectstorage.CsvObjectStorageCachedContentSourceRepository {
	cacheRefreshInterval, err := time.ParseDuration(utils.GetEnvOrDefault(static.ContentSourceCacheRefreshInterval))
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"func": "createCsvObjectStorageCachedContentSourceRepository",
			"err":  err,
			"key":  static.ContentSourceCacheRefreshInterval.Key,
		}).Fatal("parsing of env variable is failed")
	}

	return objectstorage.NewCsvObjectStorageCachedContentSourceRepository(
		getObjectStorageClient(),
		utils.MustGetEnv(static.YandexObjectStorageContentSourceBucket),
		utils.MustGetEnv(static.YandexObjectStorageContentSourceBucketKey),
		cacheRefreshInterval,
	)
}

func createPostgresqlCachedContentSourceRepository() repository.ContentSourceRepository {
	pgContentSourceCacheRefreshInterval, err := time.ParseDuration(utils.GetEnvOrDefault(static.ContentSourceCacheRefreshInterval))
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"func": "createPostgresqlCachedContentSourceRepository",
			"err":  err,
			"key":  static.ContentSourceCacheRefreshInterval.Key,
		}).Fatal("parsing of env variable is failed")
	}

	return pg.NewCachedPgContentSourceRepository(getPostgresqlConnection(), pgContentSourceCacheRefreshInterval)
}

func CreateMembershipWarningRepository(repoType RepositoryType) repository.MembershipWarningRepository {
	var repo repository.MembershipWarningRepository
	switch repoType {
	case CsvYandexObjectStorage:
		repo = createCsvObjectStorageMembershipWarningRepository()
	case Postgresql:
		fallthrough
	default:
		repo = pg.NewPgMembershipWarningRepository(getPostgresqlConnection())
	}

	return repo
}

func createCsvObjectStorageMembershipWarningRepository() *objectstorage.CsvObjectStorageMembershipWarningRepository {
	return objectstorage.NewCsvObjectStorageMembershipWarningRepository(
		getObjectStorageClient(),
		utils.MustGetEnv(static.YandexObjectStorageMembershipWarningBucket),
	)
}
