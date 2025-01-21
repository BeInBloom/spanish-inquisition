package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	app "github.com/BeInBloom/spanish-inquisition/internal/app/server-app"
	config "github.com/BeInBloom/spanish-inquisition/internal/config/server-config"
	"github.com/BeInBloom/spanish-inquisition/internal/logger"
)

func main() {
	fmt.Printf("Read config...\n")
	cfg := config.New()

	logger, err := logger.New(cfg.Env)
	if err != nil {
		panic(err)
	}

	logger.Info("Logger initialized")

	logger.Info("Starting server...")
	app := app.New(cfg.ServerConfig, logger)
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
