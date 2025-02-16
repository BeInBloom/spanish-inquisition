package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"syscall"
	"time"

	config "github.com/BeInBloom/spanish-inquisition/internal/config/server-config"
	"github.com/BeInBloom/spanish-inquisition/internal/handlers"
	"github.com/BeInBloom/spanish-inquisition/internal/middlewares"
	"github.com/BeInBloom/spanish-inquisition/internal/models"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

const (
	Counter   = "counter"
	Gauge     = "gauge"
	UpdateURL = "/update/"
	RootURL   = "/"
)

type repository interface {
	CreateOrUpdate(models.Metrics) error
	Get(models.Metrics) (models.Metrics, error)
	Dump() ([]models.Metrics, error)
	Check() error
}

type app struct {
	server *http.Server
	repo   repository
	log    *zap.Logger
}

func New(config config.ServerConfig, log *zap.Logger, repo repository) *app {
	return &app{
		server: &http.Server{
			Addr:         config.Address,
			Handler:      nil,
			ReadTimeout:  config.Timeout,
			WriteTimeout: config.Timeout,
			IdleTimeout:  config.IdleTimeout,
		},
		repo: repo,
		log:  log,
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
		if closeErr := a.server.Close(); closeErr != nil {
			return fmt.Errorf("shutdown error: %v, close error: %v", err, closeErr)
		}
		return err
	}

	if err := a.log.Sync(); err != nil {
		if !errors.Is(err, syscall.EINVAL) {
			return err
		}
	}

	return nil
}

func (a *app) Init() {
	a.initHandlers()
}

func (a *app) initHandlers() {
	r := chi.NewRouter()

	r.Use(
		middleware.Compress(5, "application/json", "text/html"),
		middlewares.Decomp,

		middleware.RequestID,
		middleware.RealIP,
		middlewares.Logger(a.log.Sugar()),
		middleware.Recoverer,
	)

	r.Route("/", func(r chi.Router) {
		r.Get("/", handlers.GetRoot(a.repo))
		r.Route("/ping", func(r chi.Router) {
			r.Get("/", handlers.Ping(a.repo))
		})
		r.Route("/value", func(r chi.Router) {
			r.With(middleware.AllowContentType("application/json")).Get("/", handlers.GetDataByJSON(a.repo))
			r.With(middleware.AllowContentType("application/json")).Post("/", handlers.GetDataByJSON(a.repo))
			r.With(middleware.AllowContentType("text/plain")).Get("/{type}/{name}/", handlers.GetData(a.repo))
		})
		r.Route("/update", func(r chi.Router) {
			r.With(middleware.AllowContentType("application/json")).Post("/", handlers.CreateOrUpdateByJSON(a.repo))
			r.With(middleware.AllowContentType("text/plain")).Post("/{type}/{name}/{value}/", handlers.CreateOrUpdate(a.repo))
		})
	})

	a.server.Handler = r
}
