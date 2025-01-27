package handlers

import (
	"net/http"
	"text/template"

	ptypes "github.com/BeInBloom/spanish-inquisition/internal/types"
)

const (
	templatePath = "./internal/template/index.html"
)

type dumper interface {
	Dump() []ptypes.Metrics
}

func GetRoot(repo dumper) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")

		metrics := repo.Dump()

		tmpl, err := template.ParseFiles(templatePath)
		if err != nil {
			http.Error(w, "Error loading template", http.StatusInternalServerError)
			return
		}

		if err := tmpl.Execute(w, metrics); err != nil {
			http.Error(w, "Error rendering template", http.StatusInternalServerError)
		}

		w.WriteHeader(http.StatusOK)
	}
}
