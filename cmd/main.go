package main

import (
	"chattweiler/pkg/bot"
	"chattweiler/pkg/repository/factory"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

var packageLogFields = logrus.Fields{
	"package": "main",
}

func main() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	pgCachedContentSourceRepository := factory.CreateContentSourceRepository(factory.Postgresql)
	pgCachedPhraseRepository := factory.CreatePhraseRepository(factory.Postgresql)
	pgMembershipWarningRepository := factory.CreateMembershipWarningRepository(factory.Postgresql)
	bot.NewBot(pgCachedPhraseRepository, pgMembershipWarningRepository, pgCachedContentSourceRepository).Start()
}
