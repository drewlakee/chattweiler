package main

import (
	"chattweiler/pkg/app/utils"
	"chattweiler/pkg/bot"
	"chattweiler/pkg/repository/pg"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"time"
)

func main() {
	vkBotToken := utils.GetEnvOrDefault("vk.community.bot.token", "unset")
	if vkBotToken == "unset" {
		logrus.WithFields(logrus.Fields{
			"package": "main",
			"func":    "main",
		}).Fatal("vk.community.bot.token is unset")
	}

	pgDataSourceString := utils.GetEnvOrDefault("pg.datasource.string", "unset")
	if pgDataSourceString == "unset" {
		logrus.WithFields(logrus.Fields{
			"package": "main",
			"func":    "main",
		}).Fatal("pg.datasource.string is unset")
	}

	db, err := sqlx.Connect("postgres", pgDataSourceString)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"package": "main",
			"func":    "main",
			"err":     err,
		}).Fatal("Postgres connection error")
	}

	pgCacheRefreshInterval, err := time.ParseDuration(utils.GetEnvOrDefault("pg.phrases.cache.refresh.interval", "15m"))
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"package": "main",
			"func":    "main",
			"err":     err,
		}).Fatal("pg.phrases.cache.refresh.interval parse error")
	}

	pgCachedPgPhraseRepository := pg.NewCachedPgPhraseRepository(db, pgCacheRefreshInterval)
	pgMembershipWarningRepository := pg.NewPgMembershipWarningRepository(db)
	bot.NewBot(vkBotToken, pgCachedPgPhraseRepository, pgMembershipWarningRepository).Start()
}
