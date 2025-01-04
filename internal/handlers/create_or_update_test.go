package handlers_test

import (
	"testing"
)

type MockSaver struct {
	data  map[string]string
	error error
}

func (m *MockSaver) CreateOrUpdate(repoID, id, item string) error {
	if m.error != nil {
		return m.error
	}
	m.data[repoID+":"+id] = item
	return nil
}

func (m *MockSaver) Dump() string {
	return "Mock dump output"
}

func TestCreateOrUpdate_Success(t *testing.T) {
}

func TestCreateOrUpdate_BadRequest(t *testing.T) {
}

func TestCreateOrUpdate_RepoNotFound(t *testing.T) {
}
