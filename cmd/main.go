package main

import (
	"chattweiler/pkg/bot"
	"chattweiler/pkg/configs"
	"chattweiler/pkg/repository"
	"chattweiler/pkg/utils"

	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

func main() {
	logrus.SetFormatter(&logrus.JSONFormatter{})

	var createRepositoryType repository.RepositoryType
	if utils.GetEnvOrDefault(configs.PgDatasourceString) != "" {
		createRepositoryType = repository.Postgresql
	} else {
		createRepositoryType = repository.CsvYandexObjectStorage
	}

	bot.NewBot(
		repository.CreatePhraseRepository(createRepositoryType),
		repository.CreateMembershipWarningRepository(createRepositoryType),
		repository.CreateContentSourceRepository(createRepositoryType),
	).Start()
}
