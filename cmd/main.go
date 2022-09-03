package main

import (
	"chattweiler/pkg/bot"
	"chattweiler/pkg/configs"
	"chattweiler/pkg/repository"
	"chattweiler/pkg/utils"

	_ "github.com/lib/pq"
)

func main() {
	var createRepositoryType repository.StorageType
	if utils.GetEnvOrDefault(configs.PgDatasourceString) != "" {
		createRepositoryType = repository.Postgresql
	} else {
		createRepositoryType = repository.CsvYandexObjectStorage
	}

	bot.NewLongPoolingBot(
		repository.CreatePhraseRepository(createRepositoryType),
		repository.CreateMembershipWarningRepository(createRepositoryType),
		repository.CreateContentSourceRepository(createRepositoryType),
	).Serve()
}
