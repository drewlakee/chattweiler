package main

import (
	"chattweiler/internal/bot"
	"chattweiler/internal/repository/factory"
	_ "github.com/lib/pq"
)

func main() {
	bot.NewLongPoolingBot(
		factory.CreatePhraseRepository(factory.CsvYandexObjectStorage),
		factory.CreateMembershipWarningRepository(factory.CsvYandexObjectStorage),
		factory.CreateContentSourceRepository(factory.CsvYandexObjectStorage),
	).Serve()
}
