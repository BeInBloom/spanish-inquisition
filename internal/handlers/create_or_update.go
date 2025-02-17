package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/BeInBloom/spanish-inquisition/internal/models"
	mr "github.com/BeInBloom/spanish-inquisition/internal/repository/memrepository"
	"github.com/go-chi/chi/v5"
)

type saver interface {
	CreateOrUpdate(models.Metrics) error
}

func CreateOrUpdate(storage saver) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		m, err := models.CreateMetricsByType(
			chi.URLParam(r, "type"),
			chi.URLParam(r, "name"),
			chi.URLParam(r, "value"),
		)
		if err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		if err := storage.CreateOrUpdate(m); err != nil {
			if errors.Is(err, mr.ErrRepoNotFound) {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}

			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)

		w.Write([]byte("ok"))
	}
}

func CreateOrUpdateByJSON(storage saver) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var data models.Metrics

		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&data); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		if err := storage.CreateOrUpdate(data); err != nil {
			if errors.Is(err, mr.ErrRepoNotFound) {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}

			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)

		w.Write([]byte("{\"status\": \"ok\"}"))
	}
}

func CreateOrUpdateByJSONBatch(storage saver) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var data []models.Metrics

		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&data); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		for _, d := range data {
			if err := storage.CreateOrUpdate(d); err != nil {
				if errors.Is(err, mr.ErrRepoNotFound) {
					http.Error(w, "not found", http.StatusNotFound)
					return
				}

				http.Error(w, "bad request", http.StatusBadRequest)
				return
			}
		}

		w.WriteHeader(http.StatusOK)

		w.Write([]byte("{\"status\": \"ok\"}"))
	}
}
