package handlers

import (
	"net/http"
	"text/template"

	ptypes "github.com/BeInBloom/spanish-inquisition/internal/types"
)

const (
	templatePath = "./internal/template/index.html"
)

type asdfa interface {
	Dump() []ptypes.Metrics
}

func GetRoot(repo asdfa) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		metrics := repo.Dump()
		tmpl, err := template.ParseFiles(templatePath)
		if err != nil {
			http.Error(w, "Error loading template", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html")
		if err := tmpl.Execute(w, metrics); err != nil {
			http.Error(w, "Error rendering template", http.StatusInternalServerError)
		}
	}
}
