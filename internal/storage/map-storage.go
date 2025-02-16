package mapstorage

import (
	"errors"
	"sync"

	"github.com/BeInBloom/spanish-inquisition/internal/models"
)

var (
	ErrNotFound             = errors.New("not found")
	ErrUnexpectedMetricType = errors.New("unexpected metric type")
)

type storage struct {
	mutex sync.Mutex
	data  map[string]models.Metrics
}

func (s *storage) Get(item models.Metrics) (models.Metrics, error) {
	const fn = "storage.Get"

	s.mutex.Lock()
	defer s.mutex.Unlock()

	key := s.getKey(item)

	value, ok := s.data[key]
	if !ok {
		return models.Metrics{}, ErrNotFound
	}

	return value, nil
}

func (s *storage) Dump() []models.Metrics {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	var result []models.Metrics

	for _, item := range s.data {
		result = append(result, item)

	}

	return result
}

func (s *storage) Create(item models.Metrics) error {
	const fn = "storage.Create"

	switch item.MType {
	case models.Counter:
		s.createCounter(item)
		return nil
	case models.Gauge:
		s.createGauge(item)
		return nil
	default:
		return ErrUnexpectedMetricType
	}

}

func (s *storage) createGauge(item models.Metrics) {
	const fn = "storage.createGauge"

	s.mutex.Lock()
	defer s.mutex.Unlock()

	key := s.getKey(item)
	s.data[key] = item
}

func (s *storage) createCounter(item models.Metrics) {
	const fn = "storage.createCounter"

	s.mutex.Lock()
	defer s.mutex.Unlock()

	key := s.getKey(item)

	_, ok := s.data[key]
	if !ok {
		s.data[s.getKey(item)] = item
		return
	}

	if s.data[key].Delta != nil && item.Delta != nil {
		*s.data[key].Delta += *item.Delta
	}
}

func (s *storage) getKey(item models.Metrics) string {
	return item.MType + item.ID
}

func New() *storage {
	data := make(map[string]models.Metrics)

	return &storage{
		data: data,
	}
}
