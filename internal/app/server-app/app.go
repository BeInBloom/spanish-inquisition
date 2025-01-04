package app

import (
	"context"
	"fmt"
	"net/http"
	"time"

	config "github.com/BeInBloom/spanish-inquisition/internal/config/server-config"
	"github.com/BeInBloom/spanish-inquisition/internal/handlers"
	"github.com/BeInBloom/spanish-inquisition/internal/repository/memrepository"
	mapstorage "github.com/BeInBloom/spanish-inquisition/internal/storage"
)

const (
	Counter   = "counter"
	Gauge     = "gauge"
	UpdateURL = "/update/"
	RootURL   = "/"
)

type repository interface {
	CreateOrUpdate(repoID string, id string, item string) error
	Get(repoID string, id string) (string, error)
	AddStorage(id string, repo any) error
	Dump() string
}

type app struct {
	server *http.Server
	repo   repository
}

func New(config config.ServerConfig) *app {
	return &app{
		server: &http.Server{
			Addr:         fmt.Sprintf("%v:%v", config.Address, config.Port),
			Handler:      nil,
			ReadTimeout:  config.Timeout,
			WriteTimeout: config.Timeout,
			IdleTimeout:  config.IdleTimeout,
		},
		repo: nil,
	}
}

func (a *app) Run() error {
	if err := a.server.ListenAndServe(); err != nil {
		return err
	}

	return nil
}

func (a *app) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()

	if err := a.server.Shutdown(ctx); err != nil {
		return a.server.Close()
	}

	return nil
}

func (a *app) Init() {
	a.initRepo()
	a.initHandlers()
}

func (a *app) initRepo() {
	counterStorage := mapstorage.NewCounterStorage()
	gaugeStorage := mapstorage.NewCommonStorage[memrepository.Gauge]()

	if a.repo == nil {
		a.repo = memrepository.New()
	}

	a.repo.AddStorage(Counter, counterStorage)
	a.repo.AddStorage(Gauge, gaugeStorage)
}

func (a *app) initHandlers() {
	mux := http.NewServeMux()

	mux.HandleFunc(RootURL, http.NotFound)
	mux.HandleFunc(UpdateURL, handlers.CreateOrUpdate(a.repo))
	//Я так и не понял, какого ляду это так работает если честно
	mux.HandleFunc("/update", http.NotFound)

	a.server.Handler = mux
}
