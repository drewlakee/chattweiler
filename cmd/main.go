package main

import (
	"chattweiler/pkg/app/utils"
	"chattweiler/pkg/bot"
	"chattweiler/pkg/repository/factory"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

func main() {
	logrus.SetFormatter(&logrus.JSONFormatter{})

	var createRepositoryType factory.RepositoryType
	if utils.GetEnvOrDefault("pg.datasource.string", "unset") != "unset" {
		createRepositoryType = factory.Postgresql
	} else {
		createRepositoryType = factory.CsvYandexObjectStorage
	}

	pgCachedContentSourceRepository := factory.CreateContentSourceRepository(createRepositoryType)
	pgCachedPhraseRepository := factory.CreatePhraseRepository(createRepositoryType)
	pgMembershipWarningRepository := factory.CreateMembershipWarningRepository(createRepositoryType)
	bot.NewBot(pgCachedPhraseRepository, pgMembershipWarningRepository, pgCachedContentSourceRepository).Start()
}
