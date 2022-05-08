package factory

import (
	"chattweiler/pkg/app/utils"
	"chattweiler/pkg/repository"
	"chattweiler/pkg/repository/pg"
	objectstorage "chattweiler/pkg/repository/yandex/s3"
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"time"
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
	pgDataSourceString := utils.GetEnvOrDefault("pg.datasource.string", "unset")
	if pgDataSourceString == "unset" {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"func": "getPostgresqlConnection",
		}).Fatal("pg.datasource.string is unset")
	}

	if pgConnectionSingleton != nil {
		return pgConnectionSingleton
	}

	pgConnectionSingleton, err := sqlx.Connect("postgres", pgDataSourceString)
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(packageLogFields).WithFields(logrus.Fields{
			"func": "getPostgresqlConnection",
			"err":  err,
		}).Fatal("Postgres connection error")
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

	yandexObjectStorageAccessKeyID := utils.GetEnvOrDefault("yandex.object.storage.access.key.id", "unset")
	if yandexObjectStorageAccessKeyID == "unset" {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"func": "getObjectStorageClient",
		}).Fatal("yandex.object.storage.access.key.id is unset")
	}

	yandexObjectStorageSecretAccessKey := utils.GetEnvOrDefault("yandex.object.storage.secret.access.key", "unset")
	if yandexObjectStorageSecretAccessKey == "unset" {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"func": "getObjectStorageClient",
		}).Fatal("yandex.object.storage.secret.access.key is unset")
	}

	credentialsProvider := aws.CredentialsProviderFunc(func(ctx context.Context) (aws.Credentials, error) {
		return aws.Credentials{
			AccessKeyID:     yandexObjectStorageAccessKeyID,
			SecretAccessKey: yandexObjectStorageSecretAccessKey,
		}, nil
	})

	yandexObjectStorageRegion := utils.GetEnvOrDefault("yandex.object.storage.region", "unset")
	if yandexObjectStorageRegion == "unset" {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"func": "getObjectStorageClient",
		}).Fatal("yandex.object.storage.region is unset")
	}

	cfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithEndpointResolverWithOptions(customResolver),
		config.WithRegion(yandexObjectStorageRegion),
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
	cacheRefreshInterval, err := time.ParseDuration(utils.GetEnvOrDefault("yandex.object.storage.phrases.cache.refresh.interval", "15m"))
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"func": "createCsvObjectStorageCachedPhraseRepository",
			"err":  err,
		}).Fatal("yandex.object.storage.phrases.cache.refresh.interval is unset or parsing error")
	}

	bucket := utils.GetEnvOrDefault("yandex.object.storage.phrases.bucket", "unset")
	if bucket == "unset" {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"func": "createCsvObjectStorageCachedPhraseRepository",
			"err":  err,
		}).Fatal("yandex.object.storage.phrases.bucket is unset or parsing error")
	}

	key := utils.GetEnvOrDefault("yandex.object.storage.phrases.bucket.key", "unset")
	if key == "unset" {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"func": "createCsvObjectStorageCachedPhraseRepository",
			"err":  err,
		}).Fatal("yandex.object.storage.phrases.bucket.key is unset or parsing error")
	}

	return objectstorage.NewCsvObjectStorageCachedPhraseRepository(
		getObjectStorageClient(),
		bucket,
		key,
		cacheRefreshInterval,
	)
}

func createPostgresqlCachedPhraseRepository() repository.PhraseRepository {
	pgPhrasesCacheRefreshInterval, err := time.ParseDuration(utils.GetEnvOrDefault("pg.phrases.cache.refresh.interval", "15m"))
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"func": "createPostgresqlCachedPhraseRepository",
			"err":  err,
		}).Fatal("pg.phrases.cache.refresh.interval parse error")
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
	cacheRefreshInterval, err := time.ParseDuration(utils.GetEnvOrDefault("yandex.object.storage.content.source.cache.refresh.interval", "15m"))
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"func": "createCsvObjectStorageCachedContentSourceRepository",
			"err":  err,
		}).Fatal("yandex.object.storage.content.source.cache.refresh.interval is unset or parsing error")
	}

	bucket := utils.GetEnvOrDefault("yandex.object.storage.phrases.bucket", "unset")
	if bucket == "unset" {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"func": "createCsvObjectStorageCachedContentSourceRepository",
			"err":  err,
		}).Fatal("yandex.object.storage.content.source.bucket is unset or parsing error")
	}

	key := utils.GetEnvOrDefault("yandex.object.storage.content.source.bucket.key", "unset")
	if key == "unset" {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"func": "createCsvObjectStorageCachedContentSourceRepository",
			"err":  err,
		}).Fatal("yandex.object.storage.content.source.bucket.key is unset or parsing error")
	}

	return objectstorage.NewCsvObjectStorageCachedContentSourceRepository(
		getObjectStorageClient(),
		bucket,
		key,
		cacheRefreshInterval,
	)
}

func createPostgresqlCachedContentSourceRepository() repository.ContentSourceRepository {
	pgContentSourceCacheRefreshInterval, err := time.ParseDuration(utils.GetEnvOrDefault("pg.content.source.cache.refresh.interval", "15m"))
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"func": "createPostgresqlCachedContentSourceRepository",
			"err":  err,
		}).Fatal("pg.content.source.cache.refresh.interval parse error")
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
	bucket := utils.GetEnvOrDefault("yandex.object.storage.membership.warning.bucket", "unset")
	if bucket == "unset" {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"func": "createCsvObjectStorageMembershipWarningRepository",
		}).Fatal("yandex.object.storage.membership.warning.bucket is unset or parsing error")
	}

	return objectstorage.NewCsvObjectStorageMembershipWarningRepository(
		getObjectStorageClient(),
		bucket,
	)
}
