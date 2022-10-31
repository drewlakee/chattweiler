package main

import (
	"chattweiler/internal/bot"
	"chattweiler/internal/configs"
	"chattweiler/internal/logging"
	"chattweiler/internal/repository"
	"chattweiler/internal/repository/factory"
	"chattweiler/internal/utils"
	_ "github.com/lib/pq"
)

func main() {
	logging.Log.Info("main", "main", "preparing bot instance...")
	logging.Log.Info("main", "main", "creating and checking phrases repository...")
	phrases := factory.CreatePhraseRepository(factory.CsvYandexObjectStorage)

	var membershipWarnings repository.MembershipWarningRepository
	if utils.GetEnvOrDefault(configs.BotFunctionalityMembershipChecking) == "true" {
		logging.Log.Info("main", "main", "creating and checking membership warnings repository...")
		membershipWarnings = factory.CreateMembershipWarningRepository(factory.CsvYandexObjectStorage)
	} else {
		membershipWarnings = nil
	}

	var commands repository.ContentCommandRepository
	if utils.MustGetEnv(configs.BotFunctionalityContentCommands) == "true" {
		logging.Log.Info("main", "main", "creating and checking commands repository...")
		commands = factory.CreateContentSourceRepository(factory.CsvYandexObjectStorage)
	} else {
		commands = nil
	}

	logging.Log.Info("main", "main", "creating bot instance...")
	bot.NewLongPoolingBot(phrases, membershipWarnings, commands).Serve()
}
