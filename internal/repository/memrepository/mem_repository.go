package memrepository

import (
	"context"
	"errors"
	"fmt"
	"time"

	config "github.com/BeInBloom/spanish-inquisition/internal/config/server-config"
	"github.com/BeInBloom/spanish-inquisition/internal/models"
	mapstorage "github.com/BeInBloom/spanish-inquisition/internal/storage"
	ptypes "github.com/BeInBloom/spanish-inquisition/internal/types"
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

type repository[T any] interface {
	Create(id string, item T) error
	Get(id string) (T, error)
	Dump() []ptypes.Metric
}

type Gauge = float64
type Counter = int64

type backuper interface {
	Save([]models.Metrics) error
	Restore() ([]models.Metrics, error)
}

type memRepository struct {
	storeInterval time.Duration
	data          map[string]any
	bak           backuper
}

func (m *memRepository) AddStorage(id string, repo any) error {
	const fn = "MemStorage.AddRepository"

	switch repo.(type) {
	case repository[Gauge]:
		m.data[id] = repo
	case repository[Counter]:
		m.data[id] = repo
	default:
		return ErrRepoNotFound
	}

	return nil
}

func (m *memRepository) Get(repoID string, id string) (string, error) {
	const fn = "MemStorage.Get"

	repo, ok := m.data[repoID]
	if !ok {
		return "", ErrNotCorrectMetricType
	}

	switch repo := repo.(type) {
	case repository[Gauge]:
		return m.getGauge(repo, id)
	case repository[Counter]:
		return m.getCounter(repo, id)
	default:
		return "", ErrNotCorrectType
	}
}

func (m *memRepository) Dump() []models.Metrics {
	var result []models.Metrics

	for repoID, repo := range m.data {
		switch repo := repo.(type) {
		case repository[Gauge]:
			result = append(result, models.Metrics{
				Type:   repoID,
				Values: repo.Dump(),
			})
		case repository[Counter]:
			result = append(result, ptypes.Metrics{
				Type:   repoID,
				Values: repo.Dump(),
			})
		}
	}

	return result
}

func (m *memRepository) CreateOrUpdate(metric models.Metrics) error {
	const fn = "MemStorage.CreateOrUpdate"

	repo, ok := m.data[metric.MType]
	if !ok {
		return ErrNotCorrectMetricType
	}

	if err := m.validateMetric(metric); err != nil {
		return err
	}

	switch repo := repo.(type) {
	case repository[Gauge]:
		if err := m.createGauge(repo, metric.ID, *metric.Value); err != nil {
			return err
		}
	case repository[Counter]:
		if err := m.createCounter(repo, metric.ID, *metric.Delta); err != nil {
			return err
		}
	default:
		return ErrNotCorrectType
	}

	//Да, я понимаю, что можно просто дописывать полученные даты в бек, но уже нету на это времени
	//Да, я понимаю, что буду сохранять полные состояния хранилища, хотя могу просто сохранить последовательность операций
	//которые меняют состояние хранилища. Потом просто пройтись по этим операциям и востановить исходное состояние
	//и писать json.NewEncoder. Будет время, переделаю
	if m.storeInterval < 1 {
		if err := m.backup(); err != nil {
			return fmt.Errorf("backup error: %v, %v", err, fn)
		}
	}

	return nil
}

func (m *memRepository) Close() error {
	return nil
}

func (m *memRepository) Init(ctx context.Context) {
	if m.storeInterval > 0 {
		m.startBackup(ctx)
	}
}

func (m *memRepository) getCounter(repo repository[Counter], id string) (string, error) {
	const fn = "MemStorage.getCounter"

	num, err := repo.Get(id)
	if err != nil {
		return "", fmt.Errorf("%v: %v", fn, err)
	}

	return fmt.Sprintf("%v", num), nil
}

func (m *memRepository) getGauge(repo repository[Gauge], id string) (string, error) {
	const fn = "MemStorage.getGauge"

	f, err := repo.Get(id)
	if err != nil {
		return "", fmt.Errorf("%v: %v", fn, err)
	}

	return fmt.Sprintf("%v", f), nil
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
}

func (m *memRepository) createGauge(repo repository[Gauge], id string, item float64) error {
	const fn = "MemStorage.createGauge"

	return repo.Create(id, Gauge(item))
}

func (m *memRepository) createCounter(repo repository[Counter], id string, delta int64) error {
	const fn = "MemStorage.createCounter"

	return repo.Create(id, Counter(delta))
}

func (m *memRepository) restoreFromBak() error {
	const fn = "MemStorage.restoreFromBak"

	data, err := m.bak.Restore()
	if err != nil {
		fmt.Printf("restore error: %v\n", err)
		return err
	}

	for _, item := range data {
		for _, metric := range item.Values {
			if err := m.CreateOrUpdate(item.Type, metric.Name, metric.Value); err != nil {
				return fmt.Errorf("%s: %v", fn, err)
			}
		}
	}

	return nil
}

func (m *memRepository) backup() error {
	data := m.Dump()

	if err := m.bak.Save(data); err != nil {
		return fmt.Errorf("backup error: %v", err)
	}

	return nil
}

func (m *memRepository) startBackup(ctx context.Context) error {
	if m.storeInterval == 0 {
		return nil
	}

	ticker := time.NewTicker(m.storeInterval)

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

func New(ctx context.Context, config *config.Config, bak backuper) *memRepository {
	memRepository := &memRepository{
		data:          make(map[string]any),
		storeInterval: time.Duration(config.StoreInterval) * time.Second,
		bak:           bak,
	}

	counterStorage := mapstorage.NewCounterStorage()
	gaugeStorage := mapstorage.NewCommonStorage[Gauge]()

	memRepository.AddStorage(counter, counterStorage)
	memRepository.AddStorage(gauge, gaugeStorage)

	if config.Restore {
		memRepository.restoreFromBak()
	}

	return memRepository
}
