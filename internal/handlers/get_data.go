package handlers

import (
	"net/http"

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

		w.Write([]byte(value))
	}
}
