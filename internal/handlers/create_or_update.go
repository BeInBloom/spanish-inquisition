package handlers

import (
	"errors"
	"net/http"

	mr "github.com/BeInBloom/spanish-inquisition/internal/repository/memrepository"
	"github.com/go-chi/chi/v5"
)

type saver interface {
	CreateOrUpdate(repoID string, id string, item string) error
}

func CreateOrUpdate(storage saver) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		typeID := chi.URLParam(r, "type")
		name := chi.URLParam(r, "name")
		value := chi.URLParam(r, "value")

		if err := storage.CreateOrUpdate(typeID, name, value); err != nil {
			if errors.Is(err, mr.ErrRepoNotFound) {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}

			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
	}
}
