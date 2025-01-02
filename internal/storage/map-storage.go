package mapstorage

import (
	"errors"
	"fmt"
	"sync"
)

var (
	ErrNotFound = errors.New("not found")
)

type storage[T any] struct {
	mutex sync.Mutex
	data  map[string]T
}

type counterStorage struct {
	storage[int64]
}

func (s *counterStorage) Create(id string, item int64) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.data[id] += item

	return nil
}

func (s *storage[T]) Get(id string) (T, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	value, ok := s.data[id]
	if !ok {
		return value, ErrNotFound
	}

	return value, nil
}

func (s *storage[T]) Dump() string {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	var result string

	for id, item := range s.data {
		result += fmt.Sprintf("%v: %v\n", id, item)
	}

	return result
}

func (s *storage[T]) Create(id string, item T) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.data[id] = item

	return nil
}

func NewCommonStorage[T any]() *storage[T] {
	return &storage[T]{
		data: make(map[string]T),
	}
}

func NewCounterStorage() *counterStorage {
	return &counterStorage{
		storage[int64]{
			data: make(map[string]int64),
		},
	}
}
