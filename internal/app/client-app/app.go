package app

import (
	"context"
	"fmt"
	"sync"
	"time"

	config "github.com/BeInBloom/spanish-inquisition/internal/config/client-config"
	"github.com/BeInBloom/spanish-inquisition/internal/models"
	"github.com/BeInBloom/spanish-inquisition/internal/wrappers"
)

type dataFetcher interface {
	Fetch() ([]models.Metrics, error)
}

type saver interface {
	Save(...models.Metrics) error
}

type app struct {
	ctx context.Context
	// client         *http.Client
	saver          saver
	fetcher        []dataFetcher
	reportInterval int64
}

func (a *app) Init(ctx context.Context) {
	a.ctx = ctx
}

func (a *app) Run() error {
	const fn = "app.Run"

	fmt.Println("Sending data...")
	if err := a.sendData(); err != nil {
		fmt.Printf("Error sending data: %v\n", err)
	}

	ticker := time.NewTicker(time.Duration(a.reportInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-a.ctx.Done():
			return a.ctx.Err()
		case <-ticker.C:
			fmt.Println("Sending data...")
			if err := wrappers.RetryWrapper(a.sendData, 3, 2*time.Second); err != nil {
				fmt.Printf("Error sending data: %v\n", err)
			}
		}
	}
}

func (a *app) Close() error {
	a.ctx.Done()
	return nil
}

func (a *app) sendData() error {
	const fn = "app.sendData"
	var mutex sync.Mutex

	var wg sync.WaitGroup
	var allData []models.Metrics
	var allErrors []error
	for _, fetcher := range a.fetcher {
		wg.Add(1)

		go func() {
			defer wg.Done()
			data, err := fetcher.Fetch()
			if err != nil {
				allErrors = append(allErrors, err)
				return
			}

			mutex.Lock()
			defer mutex.Unlock()
			allData = append(allData, data...)
		}()

	}

	if err := a.saver.Save(allData...); err != nil {
		return fmt.Errorf("%s: %v", fn, err)
	}

	return nil
}

func (a *app) AddFetcher(fetcher ...dataFetcher) {
	a.fetcher = append(a.fetcher, fetcher...)
}

func New(saver saver, config config.AppConfig) *app {
	return &app{
		saver:          saver,
		fetcher:        nil,
		reportInterval: int64(config.ReportInterval),
	}
}
