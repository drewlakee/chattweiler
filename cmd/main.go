package main

import (
	"chattweiler/pkg/app/utils"
	"chattweiler/pkg/bot"
	"chattweiler/pkg/repository/pg"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"time"
)

func main() {
	vkBotToken := utils.GetEnvOrDefault("vk.community.bot.token", "unset")
	if vkBotToken == "unset" {
		panic(errors.New("vk.community.bot.token is unset"))
	}

	pgDataSourceString := utils.GetEnvOrDefault("pg.datasource.string", "unset")
	if pgDataSourceString == "unset" {
		panic(errors.New("pg.datasource.string is unset"))
	}

	db, err := sqlx.Connect("postgres", pgDataSourceString)
	if err != nil {
		fmt.Println(err)
		return
	}

	pgCacheRefreshInterval, err := time.ParseDuration(utils.GetEnvOrDefault("pg.phrases.cache.refresh.interval", "15m"))
	pgCachedPgPhraseRepository := pg.NewCachedPgPhraseRepository(db, pgCacheRefreshInterval)
	pgMembershipWarningRepository := pg.NewPgMembershipWarningRepository(db)
	bot.NewBot(vkBotToken, pgCachedPgPhraseRepository, pgMembershipWarningRepository).Start()
}
