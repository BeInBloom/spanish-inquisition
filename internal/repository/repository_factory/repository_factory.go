package repositoryfactory

import (
	"context"

	config "github.com/BeInBloom/spanish-inquisition/internal/config/server-config"
	"github.com/BeInBloom/spanish-inquisition/internal/models"
	"github.com/BeInBloom/spanish-inquisition/internal/repository/memrepository"
	sqlrepository "github.com/BeInBloom/spanish-inquisition/internal/repository/sql_repository"
)

type repository interface {
	CreateOrUpdate(models.Metrics) error
	Get(models.Metrics) (models.Metrics, error)
	Dump() ([]models.Metrics, error)
	Check() error
	Init(context.Context) error
	Close() error
}

func NewRepository(cfg config.Config) repository {
	if cfg.DBConfig.Address != "" {
		repo, err := newSQLRepository(cfg.DBConfig)
		if err != nil {
			panic(err)
		}

		return repo
	}

	return newMapRepository(cfg)
}

func newMapRepository(cfg config.Config) repository {
	return memrepository.New(cfg)
}

func newSQLRepository(cfg config.DBConfig) (repository, error) {
	return sqlrepository.New(cfg)
}
