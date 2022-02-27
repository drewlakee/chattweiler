package main

import (
	"chattweiler/pkg/bot"
	"chattweiler/pkg/repository/pg"
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"os"
	"time"
)

func main() {
	vkBotToken := os.Getenv("vk.community.bot.token")
	pgDataSourceString := os.Getenv("pg.datasource.string")
	db, err := sqlx.Connect("postgres", pgDataSourceString)
	if err != nil {
		fmt.Println(err)
		return
	}

	rawPgCacheRefreshInterval := os.Getenv("pg.phrases.cache.refresh.interval")
	pgCacheRefreshInterval, err := time.ParseDuration(rawPgCacheRefreshInterval)

	pgCachedPgPhraseRepository := pg.NewCachedPgPhraseRepository(db, pgCacheRefreshInterval)
	pgMembershipWarningRepository := pg.NewPgMembershipWarningRepository(db)

	worker := bot.NewBot(vkBotToken, pgCachedPgPhraseRepository, pgMembershipWarningRepository)

	err = worker.Start()
	if err != nil {
		fmt.Println(err)
	}
}
