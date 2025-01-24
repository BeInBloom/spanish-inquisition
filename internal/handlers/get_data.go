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
	Get(typeID, id string) (string, error)
}

func GetData(repo fetcher) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		//Привязка к роутеру?
		typeID := chi.URLParam(r, "type")
		id := chi.URLParam(r, "name")

		value, err := repo.Get(typeID, id)
		if err != nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "text/plain")

		w.Write([]byte(value))
	}
}

func GetDataByJSON(repo fetcher) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var data models.Metrics

		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&data); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		value, err := repo.Get(data.MType, data.ID)
		if err != nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		m, err := makeGetResponse(data, value)
		if err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		if err := json.NewEncoder(w).Encode(m); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
	}
}

func makeGetResponse(m models.Metrics, val string) (models.Metrics, error) {
	switch m.MType {
	case models.Gauge:
		num, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return models.Metrics{}, err
		}

		return models.Metrics{
			ID:    m.ID,
			MType: m.MType,
			Value: &num,
		}, nil
	case models.Counter:
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
