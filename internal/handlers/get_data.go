package handlers

import (
	"crypto/sha256"
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

		fmt.Printf("GetData: %+v\n", m)

		value, err := repo.Get(m)

		fmt.Printf("GetData value: %+v\n", value)

		if err != nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(parsMetricsForValue(value)))
	}
}

func GetDataByJSON(repo fetcher, k string) func(w http.ResponseWriter, r *http.Request) {
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

		jsonString, err := json.Marshal(value)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		if k != "" {
			createSignature(jsonString, &w)
		}

		w.Write(jsonString)

		w.WriteHeader(http.StatusOK)
	}
}

func createSignature(data []byte, r *http.ResponseWriter) {
	hash := sha256.Sum256(data)
	(*r).Header().Set("HashSHA256", string(hash[:]))
}

func parsMetricsForValue(data models.Metrics) string {
	switch data.MType {
	case models.Gauge:
		return strconv.FormatFloat(*data.Value, 'f', -1, 64)
	case models.Counter:
		return fmt.Sprintf("%d", *data.Delta)
	default:
		return ""
	}
}
