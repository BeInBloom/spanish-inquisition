package memrepository

import (
	"context"
	"errors"
	"fmt"
	"time"

	config "github.com/BeInBloom/spanish-inquisition/internal/config/server-config"
	filestorage "github.com/BeInBloom/spanish-inquisition/internal/metric_storage/file_storage"
	"github.com/BeInBloom/spanish-inquisition/internal/models"
	mapstorage "github.com/BeInBloom/spanish-inquisition/internal/storage"
)

const (
	counter = "counter"
	gauge   = "gauge"
)

var (
	ErrNotCorrectType       = errors.New("not correct type")
	ErrNotCorrectMetricType = errors.New("not correct metric type")
	ErrRepoNotFound         = errors.New("repository not found")
)

type repository interface {
	Create(models.Metrics) error
	Get(models.Metrics) (models.Metrics, error)
	Dump() []models.Metrics
}

type Gauge = float64
type Counter = int64

type backuper interface {
	Save([]models.Metrics) error
	Restore() ([]models.Metrics, error)
}

type memRepository struct {
	cfg  config.Config
	data repository
	bak  backuper
}

func (m *memRepository) Get(metric models.Metrics) (models.Metrics, error) {
	const fn = "MemStorage.Get"

	if err := m.validateMetric(metric); err != nil {
		return models.Metrics{}, fmt.Errorf("%v: %v", fn, err)
	}

	return m.data.Get(metric)
}

func (m *memRepository) Dump() ([]models.Metrics, error) {
	return m.data.Dump(), nil
}

func (m *memRepository) CreateOrUpdate(metric models.Metrics) error {
	const fn = "MemStorage.CreateOrUpdate"

	if err := m.validateMetric(metric); err != nil {
		return fmt.Errorf("%v: %v", fn, err)
	}

	if err := m.data.Create(metric); err != nil {
		return fmt.Errorf("%v: %v", fn, err)
	}

	//Да, я понимаю, что можно просто дописывать полученные даты в бек, но уже нету на это времени
	//Да, я понимаю, что буду сохранять полные состояния хранилища, хотя могу просто сохранить последовательность операций
	//которые меняют состояние хранилища. Потом просто пройтись по этим операциям и востановить исходное состояние
	//и писать json.NewEncoder. Будет время, переделаю
	if time.Duration(m.cfg.StoreInterval)*time.Second < 1 {
		if err := m.backup(); err != nil {
			return fmt.Errorf("backup error: %v, %v", err, fn)
		}
	}

	return nil
}

func (m *memRepository) Close() error {
	return m.backup()
}

func (m *memRepository) Init(ctx context.Context) error {
	const fn = "MemStorage.Init"

	storage := mapstorage.New()

	m.data = storage

	bak, err := filestorage.New(m.cfg.BakConfig)
	if err != nil {
		return fmt.Errorf("%v: %v", fn, err)
	}

	m.bak = bak

	if m.cfg.Restore {
		if err := m.restoreFromBak(); err != nil {
			return fmt.Errorf("%v: %v", fn, err)
		}
	}

	if err := m.startBackup(ctx); err != nil {
		return fmt.Errorf("%v: %v", fn, err)
	}

	return nil
}

func (m *memRepository) Check() error {
	return nil
}

func (m *memRepository) validateMetric(metric models.Metrics) error {
	switch metric.MType {
	case gauge:
		if metric.Value == nil {
			return ErrNotCorrectMetricType
		}
	case counter:
		if metric.Delta == nil {
			return ErrNotCorrectMetricType
		}
	default:
		return ErrNotCorrectMetricType
	}

	return nil
}

func (m *memRepository) restoreFromBak() error {
	const fn = "MemStorage.restoreFromBak"

	data, err := m.bak.Restore()
	if err != nil {
		fmt.Printf("restore error: %v\n", err)
		return err
	}

	for _, item := range data {
		if err := m.CreateOrUpdate(item); err != nil {
			fmt.Printf("restore error: %v\n", err)
			return err
		}
	}

	return nil
}

func (m *memRepository) backup() error {
	const fn = "MemStorage.backup"

	data, _ := m.Dump()

	if err := m.bak.Save(data); err != nil {
		return fmt.Errorf("backup error: %v", err)
	}

	return nil
}

func (m *memRepository) startBackup(ctx context.Context) error {
	if m.cfg.StoreInterval == 0 {
		return nil
	}

	ticker := time.NewTicker(time.Duration(m.cfg.StoreInterval) * time.Second)

	go func() {
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				//TODO: пока так, потом переделаю
				if err := m.backup(); err != nil {
					fmt.Printf("backup error: %v\n", err)
				}
			}
		}

	}()

	return nil
}

func New(config config.Config) *memRepository {
	memRepository := &memRepository{
		cfg: config,
	}

	return memRepository
}
