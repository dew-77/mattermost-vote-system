package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/dew-77/mattermost-vote-system/internal/app"
	"github.com/dew-77/mattermost-vote-system/internal/config"
	"github.com/dew-77/mattermost-vote-system/internal/mattermost"
	"github.com/dew-77/mattermost-vote-system/internal/repository"
	"github.com/dew-77/mattermost-vote-system/pkg/logger"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic("Failed to load config: " + err.Error())
	}
	
	log := logger.NewLogger(cfg.Bot.LogLevel)
	
	repo, err := repository.NewTarantoolRepository(&cfg.Tarantool)
	if err != nil {
		log.WithError(err).Fatal("Failed to connect to Tarantool")
	}
	
	mmClient, err := mattermost.NewClient(&cfg.Mattermost)
	if err != nil {
		log.WithError(err).Fatal("Failed to connect to Mattermost")
	}
	
	application := app.NewApp(cfg, log, mmClient, repo)
	
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	
	go func() {
		<-c
		log.Info("Shutting down...")
		os.Exit(0)
	}()
	
	// Запускаем бота
	log.Info("Starting bot...")
	if err := application.Start(); err != nil {
		log.WithError(err).Fatal("Failed to start bot")
	}
}