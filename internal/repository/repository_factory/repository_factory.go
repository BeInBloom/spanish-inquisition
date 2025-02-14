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
	Get(models.Metrics) (string, error)
	Dump() ([]models.Metrics, error)
	Check() error
	Init(context.Context) error
	Close() error
}

type backuper interface {
	Save([]models.Metrics) error
	Restore() ([]models.Metrics, error)
}

func NewRepository(cfg config.Config) repository {
	repo, err := newSQLRepository(cfg.DBConfig)
	if err != nil {
		return newMapRepository(cfg)
	}

	return repo
}

func newMapRepository(cfg config.Config) repository {
	return memrepository.New(cfg)
}

func newSQLRepository(cfg config.DBConfig) (repository, error) {
	return sqlrepository.New(cfg)
}
