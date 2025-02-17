package sqlrepository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	config "github.com/BeInBloom/spanish-inquisition/internal/config/server-config"
	"github.com/BeInBloom/spanish-inquisition/internal/models"
	sq "github.com/Masterminds/squirrel"
	_ "github.com/jackc/pgx/v5/stdlib"
)

var (
	ErrCantOpenDB           = errors.New("can't open db")
	ErrNotCorrectType       = errors.New("not correct type")
	ErrNotCorrectMetricType = errors.New("not correct metric type")
	ErrRepoNotFound         = errors.New("repository not found")
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
	const fn = "sqlRepository.Dump"

	query := sq.Select("id", "type", "delta", "value").
		From("metric")

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("%v: %v", fn, err)
	}

	rows, err := r.db.Query(sqlQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("%v: %v", fn, err)
	}

	defer rows.Close()

	var res []models.Metrics
	for rows.Next() {
		var m models.Metrics

		if err := rows.Scan(&m.ID, &m.MType, &m.Delta, &m.Value); err != nil {
			return nil, fmt.Errorf("%v: %v", fn, err)
		}

		res = append(res, m)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%v: %v", fn, err)
	}

	return res, nil
}

func (r *sqlRepository) Get(m models.Metrics) (models.Metrics, error) {
	const fn = "sqlRepository.Get"

	query := sq.Select("id", "type", "delta", "value").
		From("metric").
		Where(sq.Eq{"id": m.ID, "type": m.MType})

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return models.Metrics{}, fmt.Errorf("%v: %v", fn, err)
	}

	var res models.Metrics
	err = r.db.QueryRow(sqlQuery, args...).Scan(
		&res.ID, &res.MType, &res.Delta, &res.Value)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Metrics{}, fmt.Errorf("%v: %v", fn, ErrRepoNotFound)
		}

		return models.Metrics{}, fmt.Errorf("%v: %v", fn, err)
	}

	return res, nil

}

func (r *sqlRepository) CreateOrUpdate(m models.Metrics) error {
	const query = `
        INSERT INTO metric (id, type, delta, value)
        VALUES ($1, $2, $3, $4)
        ON CONFLICT (id, type) DO UPDATE SET
            delta = CASE
                WHEN metric.type = 'counter' THEN COALESCE(metric.delta, 0) + COALESCE(EXCLUDED.delta, 0)
                ELSE EXCLUDED.delta
            END,
            value = CASE
                WHEN metric.type = 'gauge' THEN EXCLUDED.value
                ELSE metric.value
            END;
    `

	if err := r.validateMetric(m); err != nil {
		return fmt.Errorf("failed to create or update metric: %w", err)
	}

	var delta sql.NullInt64
	var value sql.NullFloat64

	if m.Delta != nil {
		delta = sql.NullInt64{Int64: *m.Delta, Valid: true}
	}
	if m.Value != nil {
		value = sql.NullFloat64{Float64: *m.Value, Valid: true}
	}

	_, err := r.db.Exec(query, m.ID, m.MType, delta, value)
	if err != nil {
		fmt.Printf("failed to create or update metric: %v\n", err)
		return fmt.Errorf("failed to create or update metric: %w", err)
	}

	return nil
}

// func (r *sqlRepository) CreateOrUpdate(m models.Metrics) error {
// 	const fn = "sqlRepository.CreateOrUpdate"

// 	if err := r.validateMetric(m); err != nil {
// 		return fmt.Errorf("%v: %v", fn, err)
// 	}

// 	var deltaValue sql.NullInt64
// 	var floatValue sql.NullFloat64

// 	if err := deltaValue.Scan(m.Delta); err != nil {
// 		return fmt.Errorf("%v: %v", fn, err)
// 	}
// 	if err := floatValue.Scan(m.Value); err != nil {
// 		return fmt.Errorf("%v: %v", fn, err)
// 	}

// 	query := sq.Insert("metric").
// 		Columns("id", "type", "delta", "value").
// 		Values(m.ID, m.MType, deltaValue, floatValue).
// 		Suffix(`ON CONFLICT (id, type) DO UPDATE SET
// 			delta = EXCLUDED.delta,
// 			value = EXCLUDED.value`)

// 	sqlQuery, args, err := query.ToSql()
// 	if err != nil {
// 		return fmt.Errorf("%v: %v", fn, err)
// 	}

// 	_, err = r.db.Exec(sqlQuery, args...)
// 	if err != nil {
// 		return fmt.Errorf("%v: %v", fn, err)
// 	}

// 	return nil
// }

func (r *sqlRepository) Init(ctx context.Context) error {
	query := `
    CREATE TABLE IF NOT EXISTS metric (
        id VARCHAR(255) NOT NULL,
        type VARCHAR(7) NOT NULL CHECK (type IN ('gauge', 'counter')),
        delta BIGINT,
        value DOUBLE PRECISION,
        PRIMARY KEY (id, type)
    );`

	_, err := r.db.ExecContext(ctx, query)
	if err != nil {
		panic(err)
	}

	return nil
}

func (r *sqlRepository) validateMetric(metric models.Metrics) error {
	switch metric.MType {
	case models.Gauge:
		if metric.Value == nil {
			return ErrNotCorrectMetricType
		}
	case models.Counter:
		if metric.Delta == nil {
			return ErrNotCorrectMetricType
		}
	default:
		return ErrNotCorrectMetricType
	}

	return nil
}
