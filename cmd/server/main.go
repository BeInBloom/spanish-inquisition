package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	app "github.com/BeInBloom/spanish-inquisition/internal/app/server-app"
	config "github.com/BeInBloom/spanish-inquisition/internal/config/server-config"
)

func main() {
	fmt.Printf("Read config...\n")
	cfg := config.New()

	fmt.Printf("Creating app...\n")
	app := app.New(cfg.ServerConfig)
	app.Init()

	fmt.Printf("Starting server...\n")
	go app.Run()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	if err := app.Close(); err != nil {
		fmt.Printf("Error during server shutdown: %v", err)
	}

	fmt.Println("Server stopped")
}
