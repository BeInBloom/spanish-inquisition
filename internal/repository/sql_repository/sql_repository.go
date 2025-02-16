package sqlrepository

import (
	"context"
	"database/sql"
	"errors"

	config "github.com/BeInBloom/spanish-inquisition/internal/config/server-config"
	"github.com/BeInBloom/spanish-inquisition/internal/models"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type repository interface {
	CreateOrUpdate(models.Metrics) error
	Get(models.Metrics) (string, error)
	Dump() ([]models.Metrics, error)
	Check() error
}

var (
	ErrCantOpenDB = errors.New("can't open db")
)

type sqlRepository struct {
	db *sql.DB
}

func New(cfg config.DBConfig) (*sqlRepository, error) {
	db, err := sql.Open(cfg.DriverName, cfg.Address)
	if err != nil {
		return nil, errors.Join(ErrCantOpenDB, err)
	}

	if err := db.Ping(); err != nil {
		return nil, errors.Join(ErrCantOpenDB, err)
	}

	return &sqlRepository{
		db: db,
	}, nil
}

func (r *sqlRepository) Close() error {
	return r.db.Close()
}

func (r *sqlRepository) Check() error {
	return r.db.Ping()
}

func (r *sqlRepository) Dump() ([]models.Metrics, error) {
	panic("implement me")
}

func (r *sqlRepository) Get(m models.Metrics) (models.Metrics, error) {
	panic("implement me")
}

func (r *sqlRepository) CreateOrUpdate(m models.Metrics) error {
	panic("implement me")
}

func (r *sqlRepository) Init(ctx context.Context) error {
	return nil
}
