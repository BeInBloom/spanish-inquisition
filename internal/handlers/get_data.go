package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/BeInBloom/spanish-inquisition/internal/models"
	"github.com/go-chi/chi/v5"
)

type fetcher interface {
	Get(metric models.Metrics) (models.Metrics, error)
}

func GetData(repo fetcher) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")

		m := models.Metrics{
			MType: chi.URLParam(r, "type"),
			ID:    chi.URLParam(r, "name"),
		}

		value, err := repo.Get(m)
		if err != nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(parsMetricsForValue(value)))
	}
}

func GetDataByJSON(repo fetcher) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var data models.Metrics

		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&data); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		value, err := repo.Get(data)
		if err != nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		if err := json.NewEncoder(w).Encode(value); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func parsMetricsForValue(data models.Metrics) string {
	switch data.MType {
	case models.Gauge:
		return strconv.FormatFloat(*data.Value, 'f', 3, 64)
	case models.Counter:
		return fmt.Sprintf("%d", *data.Delta)
	default:
		return ""
	}
}
