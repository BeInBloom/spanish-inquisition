package handlers

import (
	"errors"
	"fmt"
	"net/http"

	h "github.com/BeInBloom/spanish-inquisition/internal/helpers"
	mr "github.com/BeInBloom/spanish-inquisition/internal/repository/memrepository"
)

type saver interface {
	CreateOrUpdate(repoID string, id string, item string) error
	Dump() string
}

func CreateOrUpdate(storage saver) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			contentType := r.Header.Get("Content-Type")
			if contentType != "text/plain" {
				http.Error(w, "bad request", http.StatusBadRequest)
				return
			}

			params := h.GetParams(r.URL.Path)
			if len(params) != 4 {
				http.Error(w, "not correct params", http.StatusNotFound)
				return
			}

			if err := storage.CreateOrUpdate(params[1], params[2], params[3]); err != nil {
				if errors.Is(err, mr.ErrRepoNotFound) {
					http.Error(w, err.Error(), http.StatusNotFound)
					return
				}

				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			fmt.Printf("%v\n", storage.Dump())

			fmt.Fprintf(w, "OK")
		}

		if r.Method == http.MethodGet {
			panic("not implemented")
		}
	}
}
