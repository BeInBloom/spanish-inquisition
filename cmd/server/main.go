package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	app "github.com/BeInBloom/spanish-inquisition/internal/app/server-app"
	config "github.com/BeInBloom/spanish-inquisition/internal/config/server-config"
	"github.com/BeInBloom/spanish-inquisition/internal/logger"
	filestorage "github.com/BeInBloom/spanish-inquisition/internal/metric_storage/file_storage"
	"github.com/BeInBloom/spanish-inquisition/internal/repository/memrepository"
)

func main() {
	fmt.Printf("Read config...\n")
	cfg := config.New()

	ctx := context.Background()

	logger, err := logger.New(cfg.Env)
	if err != nil {
		panic(err)
	}

	logger.Info("Logger initialized")

	logger.Info("Initializing repositories...")
	bak, err := filestorage.New(cfg.DatabaseConfig.Path)
	if err != nil {
		panic(err)
	}
	repo := memrepository.New(ctx, cfg, bak)
	logger.Info("Repositories initialized")

	logger.Info(fmt.Sprintf("Starting server on %s", cfg.ServerConfig.Address))
	app := app.New(cfg.ServerConfig, logger, repo)
	app.Init()
	logger.Info("Server initialized")

	logger.Info("Starting server...")
	go app.Run()
	logger.Info("Server started")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	if err := app.Close(); err != nil {
		logger.Error(err.Error())
	}

	logger.Info("Server stopped")
}
