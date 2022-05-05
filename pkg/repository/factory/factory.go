package factory

import (
	"chattweiler/pkg/app/utils"
	"chattweiler/pkg/repository"
	"chattweiler/pkg/repository/pg"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"time"
)

var packageLogFields = logrus.Fields{
	"package": "factory",
}

type RepositoryType string

const (
	Postgresql RepositoryType = "postgresql"
)

var pgConnectionSingleton *sqlx.DB

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

func CreatePhraseRepository(repoType RepositoryType) repository.PhraseRepository {
	var repo repository.PhraseRepository
	switch repoType {
	case Postgresql:
		fallthrough
	default:
		repo = createPostgresqlCachedPhraseRepository()
	}

	return repo
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
	case Postgresql:
		fallthrough
	default:
		repo = createPostgresqlCachedContentSourceRepository()
	}

	return repo
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
	case Postgresql:
		fallthrough
	default:
		repo = pg.NewPgMembershipWarningRepository(getPostgresqlConnection())
	}

	return repo
}
