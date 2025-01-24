package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	app "github.com/BeInBloom/spanish-inquisition/internal/app/client-app"
	datafetcher "github.com/BeInBloom/spanish-inquisition/internal/app/data-fetcher"
	config "github.com/BeInBloom/spanish-inquisition/internal/config/client-config"
	"github.com/BeInBloom/spanish-inquisition/internal/data-saver/httpsaver"
)

func main() {
	cfg := config.New()

	ctx, cancel := context.WithCancel(context.Background())

	fetcher := datafetcher.New(ctx, int64(cfg.PollInterval))
	saver := httpsaver.New(cfg.SaverConfig)

	app := app.New(fetcher, saver, cfg.AppConfig)
	app.Init(ctx)

	fmt.Println("Agent started")

	errChn := make(chan error, 1)
	go func() {
		if err := app.Run(); err != nil {
			errChn <- err
			panic(err)
		}
	}()

	sysCalls := make(chan os.Signal, 1)
	signal.Notify(sysCalls, syscall.SIGINT, syscall.SIGTERM)
	<-sysCalls

	cancel()

	fmt.Println("Agent stopped")
}
