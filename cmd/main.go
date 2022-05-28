package main

import (
	"chattweiler/pkg/app/configs/static"
	"chattweiler/pkg/app/utils"
	"chattweiler/pkg/bot"
	"chattweiler/pkg/repository/factory"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

func main() {
	logrus.SetFormatter(&logrus.JSONFormatter{})

	var createRepositoryType factory.RepositoryType
	if utils.GetEnvOrDefault(static.PgDatasourceString) != "" {
		createRepositoryType = factory.Postgresql
	} else {
		createRepositoryType = factory.CsvYandexObjectStorage
	}

	bot.NewBot(
		factory.CreatePhraseRepository(createRepositoryType),
		factory.CreateMembershipWarningRepository(createRepositoryType),
		factory.CreateContentSourceRepository(createRepositoryType),
	).Start()
}
