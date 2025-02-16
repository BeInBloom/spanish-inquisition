package mapstorage

import (
	"errors"
	"sync"

	"github.com/BeInBloom/spanish-inquisition/internal/models"
)

var (
	ErrNotFound = errors.New("not found")
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

	s.mutex.Lock()
	defer s.mutex.Unlock()

	key := s.getKey(item)

	s.data[key] = item

	return nil
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
