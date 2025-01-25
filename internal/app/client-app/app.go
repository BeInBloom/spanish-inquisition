package app

import (
	"context"
	"fmt"
	"sync"
	"time"

	config "github.com/BeInBloom/spanish-inquisition/internal/config/client-config"
	ptypes "github.com/BeInBloom/spanish-inquisition/internal/types"
)

type dataFetcher interface {
	Fetch() ([]ptypes.SendData, error)
}

type saver interface {
	Save(ptypes.SendData) error
}

type app struct {
	ctx context.Context
	// client         *http.Client
	saver          saver
	fetcher        dataFetcher
	reportInterval int64
}

func (a *app) Init(ctx context.Context) {
	a.ctx = ctx
}

func (a *app) Run() error {
	const fn = "app.Run"

	// fmt.Println("Sending data...")
	// if err := a.sendData(); err != nil {
	// 	fmt.Printf("Error sending data: %v\n", err)
	// 	return fmt.Errorf("%s: %v", fn, err)
	// }

	ticker := time.NewTicker(time.Duration(a.reportInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-a.ctx.Done():
			return a.ctx.Err()
		case <-ticker.C:
			fmt.Println("Sending data...")
			if err := a.sendData(); err != nil {
				fmt.Printf("Error sending data: %v\n", err)
				return fmt.Errorf("%s: %v", fn, err)
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

	data, err := a.fetcher.Fetch()
	if err != nil {
		return fmt.Errorf("%s: %v", fn, err)
	}

	errs := make(chan error, len(data))
	wg := &sync.WaitGroup{}

	for _, d := range data {
		wg.Add(1)

		go func(d ptypes.SendData) {
			defer wg.Done()
			if err := a.saver.Save(d); err != nil {
				errs <- err
			}
		}(d)
	}

	wg.Wait()
	close(errs)

	var errsList []error
	for err := range errs {
		errsList = append(errsList, err)
	}

	if len(errsList) > 0 {
		fmt.Printf("Errors: %v\n", errsList)
		return fmt.Errorf("%s: %v", fn, errsList)
	}

	return nil
}

func New(fetcher dataFetcher, saver saver, config config.AppConfig) *app {
	return &app{
		saver:          saver,
		fetcher:        fetcher,
		reportInterval: int64(config.ReportInterval),
	}
}
