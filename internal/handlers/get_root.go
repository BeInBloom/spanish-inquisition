package handlers

import (
	"net/http"
	"text/template"

	"github.com/BeInBloom/spanish-inquisition/internal/models"
)

const (
	templatePath = "./internal/template/index.html"
)

type dumper interface {
	Dump() ([]models.Metrics, error)
}

func GetRoot(repo dumper) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")

		metrics, err := repo.Dump()
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		tmpl, err := template.ParseFiles(templatePath)
		if err != nil {
			http.Error(w, "Error loading template", http.StatusInternalServerError)
			return
		}

		if err := tmpl.Execute(w, metrics); err != nil {
			http.Error(w, "Error rendering template", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
