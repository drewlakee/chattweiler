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

var packageLogFields = logrus.Fields{
	"package": "main",
}

func main() {
	logrus.SetFormatter(&logrus.JSONFormatter{})

	vkBotToken := utils.GetEnvOrDefault("vk.community.bot.token", "unset")
	if vkBotToken == "unset" {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"func": "main",
		}).Fatal("vk.community.bot.token is unset")
	}

	pgDataSourceString := utils.GetEnvOrDefault("pg.datasource.string", "unset")
	if pgDataSourceString == "unset" {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"func": "main",
		}).Fatal("pg.datasource.string is unset")
	}

	db, err := sqlx.Connect("postgres", pgDataSourceString)
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(packageLogFields).WithFields(logrus.Fields{
			"func": "main",
			"err":  err,
		}).Fatal("Postgres connection error")
	}

	pgPhrasesCacheRefreshInterval, err := time.ParseDuration(utils.GetEnvOrDefault("pg.phrases.cache.refresh.interval", "15m"))
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"func": "main",
			"err":  err,
		}).Fatal("pg.phrases.cache.refresh.interval parse error")
	}

	pgContentSourceCacheRefreshInterval, err := time.ParseDuration(utils.GetEnvOrDefault("pg.content.source.cache.refresh.interval", "15m"))
	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"func": "main",
			"err":  err,
		}).Fatal("pg.content.source.cache.refresh.interval parse error")
	}

	pgCachedContentSourceRepository := pg.NewCachedPgContentSourceRepository(db, pgContentSourceCacheRefreshInterval)
	pgCachedPhraseRepository := pg.NewCachedPgPhraseRepository(db, pgPhrasesCacheRefreshInterval)
	pgMembershipWarningRepository := pg.NewPgMembershipWarningRepository(db)
	bot.NewBot(vkBotToken, pgCachedPhraseRepository, pgMembershipWarningRepository, pgCachedContentSourceRepository).Start()
}
