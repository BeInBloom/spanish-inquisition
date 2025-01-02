package memrepository

import (
	"errors"
	"fmt"
	"strconv"
)

var (
	ErrNotCorrectType       = errors.New("not correct type")
	ErrNotCorrectMetricType = errors.New("not correct metric type")
	ErrRepoNotFound         = errors.New("repository not found")
)

type repository[T any] interface {
	Create(id string, item T) error
	Get(id string) (T, error)
	Dump() string
}

type Gauge = float64
type Counter = int64

type memRepository struct {
	data map[string]any
}

func (m *memRepository) AddStorage(id string, repo any) error {
	const fn = "MemStorage.AddRepository"

	switch repo.(type) {
	case repository[Gauge]:
		m.data[id] = repo
	case repository[Counter]:
		m.data[id] = repo
	default:
		fmt.Printf("repo: %v\n%v\n", repo, fn)
		return ErrRepoNotFound
	}

	return nil
}

func (m *memRepository) Get(repo_id string, id string) (string, error) {
	const fn = "MemStorage.Get"

	repo, ok := m.data[repo_id]
	if !ok {
		return "", ErrNotCorrectMetricType
	}

	switch repo := repo.(type) {
	case repository[Gauge]:
		return m.getGauge(repo, id)
	case repository[Counter]:
		return m.getCounter(repo, id)
	default:
		fmt.Printf("repo: %v\n%v\n", repo, fn)
		return "", ErrNotCorrectType
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

func (m *memRepository) CreateOrUpdate(repo_id string, id string, item string) error {
	const fn = "MemStorage.CreateOrUpdate"

	repo, ok := m.data[repo_id]
	if !ok {
		return ErrNotCorrectMetricType
	}

	switch repo := repo.(type) {
	case repository[Gauge]:
		return m.createGauge(repo, id, item)
	case repository[Counter]:
		return m.createCounter(repo, id, item)
	default:
		fmt.Printf("repo: %v\n%v\n", repo, fn)
		return ErrNotCorrectType
	}
}

func (m *memRepository) createGauge(repo repository[Gauge], id, item string) error {
	const fn = "MemStorage.createGauge"

	f, err := strconv.ParseFloat(item, 64)
	if err != nil {
		return ErrNotCorrectType
	}

	return repo.Create(id, Gauge(f))
}

func (m *memRepository) createCounter(repo repository[Counter], id, item string) error {
	const fn = "MemStorage.createCounter"

	num, err := strconv.Atoi(item)
	if err != nil {
		return ErrNotCorrectType
	}

	return repo.Create(id, Counter(num))
}

func (m *memRepository) Dump() string {
	var result string

	for repo_id, repo := range m.data {
		switch repo := repo.(type) {
		case repository[Gauge]:
			result += fmt.Sprintf("%v:\n%v\n", repo_id, repo.Dump())
		case repository[Counter]:
			result += fmt.Sprintf("%v:\n%v\n", repo_id, repo.Dump())
		}
	}

	return result
}

func New() *memRepository {
	return &memRepository{data: make(map[string]any)}
}
