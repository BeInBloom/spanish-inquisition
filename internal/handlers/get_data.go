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
	Get(metric models.Metrics) (string, error)
}

func GetData(repo fetcher) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		//Привязка к роутеру?

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
		w.Write([]byte(value))
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

		m, err := makeGetResponse(data, value)
		if err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		if err := json.NewEncoder(w).Encode(m); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func makeGetResponse(m models.Metrics, val string) (models.Metrics, error) {
	switch m.MType {
	case "gauge":
		num, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return models.Metrics{}, err
		}

		return models.Metrics{
			ID:    m.ID,
			MType: m.MType,
			Value: &num,
		}, nil
	case "counter":
		val, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return models.Metrics{}, err
		}
		return models.Metrics{
			ID:    m.ID,
			MType: m.MType,
			Delta: &val,
		}, nil
	default:
		return models.Metrics{}, fmt.Errorf("unknown metric type: %s", m.MType)
	}
}
