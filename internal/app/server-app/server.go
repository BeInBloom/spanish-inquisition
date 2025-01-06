package app

import (
	"context"
	"net/http"
	"time"

	config "github.com/BeInBloom/spanish-inquisition/internal/config/server-config"
	"github.com/BeInBloom/spanish-inquisition/internal/handlers"
	"github.com/BeInBloom/spanish-inquisition/internal/repository/memrepository"
	mapstorage "github.com/BeInBloom/spanish-inquisition/internal/storage"
	ptypes "github.com/BeInBloom/spanish-inquisition/internal/types"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
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
	Dump() []ptypes.Metrics
}

type app struct {
	server *http.Server
	repo   repository
}

func New(config config.ServerConfig) *app {
	return &app{
		server: &http.Server{
			Addr:         config.Address,
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
	r := chi.NewRouter()

	r.Use(
		middleware.RequestID,
		middleware.RealIP,
		middleware.Recoverer,
	)
	//http://<АДРЕС_СЕРВЕРА>/update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>
	//GET http://<АДРЕС_СЕРВЕРА>/value/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>
	r.Route("/", func(r chi.Router) {
		r.Get("/", handlers.GetRoot(a.repo))
		r.Get("/value/{type}/{name}", handlers.GetData(a.repo))
		r.Route("/update", func(r chi.Router) {
			r.With(middleware.AllowContentType("text/plain")).Post("/{type}/{name}/{value}", handlers.CreateOrUpdate(a.repo))
			// r.Get("/{type}/{name}", handlers.GetData(a.repo))
		})
	})

	a.server.Handler = r
}
