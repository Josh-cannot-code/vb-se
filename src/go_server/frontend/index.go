package frontend

import (
	"net/http"
	"os"
	"text/template"

	slogctx "github.com/veqryn/slog-context"
)

func Index() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get logger
		ctx := slogctx.With(r.Context(), "function", "index")
		log := slogctx.FromCtx(ctx)

		w.Header().Add("Content-Type", "text/html")

		const indexFilePath = "./frontend/templates/index.html.tmpl"
		htmlTemplate, err := template.ParseFiles(indexFilePath)
		if err != nil {
			log.Error("could not load html template from file", "error", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		const htmxFilePath = "./frontend/templates/htmx.min.js"
		data, err := os.ReadFile(htmxFilePath)
		if err != nil {
			log.Error("could not load htmx from file", "error", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		err = htmlTemplate.Execute(w, string(data))
		if err != nil {
			log.Error("could not write html to response", "error", err.Error())
		}
	})
}
