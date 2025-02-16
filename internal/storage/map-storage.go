package mapstorage

import (
	"crypto/md5"
	"errors"
	"fmt"
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

	hash, err := s.getHash(item)
	if err != nil {
		return models.Metrics{}, fmt.Errorf("%s: %v", fn, err)
	}

	value, ok := s.data[hash]
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

	hash, err := s.getHash(item)
	if err != nil {
		return fmt.Errorf("%s: %v", fn, err)
	}

	s.data[hash] = item

	return nil
}

func (s *storage) getHash(item models.Metrics) (string, error) {
	const fn = "storage.getHash"
	hash := md5.New()

	_, err := hash.Write([]byte(item.ID))
	if err != nil {
		return "", fmt.Errorf("%s: %v", fn, err)
	}

	_, err = hash.Write([]byte(item.MType))
	if err != nil {
		return "", fmt.Errorf("%s: %v", fn, err)
	}

	return string(hash.Sum(nil)), nil
}

func New() *storage {
	data := make(map[string]models.Metrics)

	return &storage{
		data: data,
	}
}
