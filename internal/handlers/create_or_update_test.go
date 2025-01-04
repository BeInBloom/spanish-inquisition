package handlers_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/BeInBloom/spanish-inquisition/internal/handlers"
	mr "github.com/BeInBloom/spanish-inquisition/internal/repository/memrepository"
	"github.com/stretchr/testify/assert"
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
	mockStorage := &MockSaver{
		data: make(map[string]string),
	}
	handler := handlers.CreateOrUpdate(mockStorage)

	reqBody := bytes.NewBufferString("test-item")
	req := httptest.NewRequest(http.MethodPost, "/api/v1/repo/123/item/456", reqBody)
	req.Header.Set("Content-Type", "text/plain")

	rec := httptest.NewRecorder()
	handler(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusNotFound, res.StatusCode)

	assert.Equal(t, "", mockStorage.data["repo:123:item:456"])
}

func TestCreateOrUpdate_BadRequest(t *testing.T) {
	mockStorage := &MockSaver{
		data: make(map[string]string),
	}
	handler := handlers.CreateOrUpdate(mockStorage)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/repo/123", nil)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	handler(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestCreateOrUpdate_RepoNotFound(t *testing.T) {
	mockStorage := &MockSaver{
		data:  make(map[string]string),
		error: mr.ErrRepoNotFound,
	}
	handler := handlers.CreateOrUpdate(mockStorage)

	reqBody := bytes.NewBufferString("test-item")
	req := httptest.NewRequest(http.MethodPost, "/api/v1/repo/123/item/456", reqBody)
	req.Header.Set("Content-Type", "text/plain")

	rec := httptest.NewRecorder()
	handler(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusNotFound, res.StatusCode)
}
