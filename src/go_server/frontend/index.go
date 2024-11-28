package frontend

import (
	"go_server/database"
	"net/http"
	"os"
	"text/template"

	slogctx "github.com/veqryn/slog-context"
)

type indexData struct {
	Videos string
	Script string
}

func Index(db database.Repository) http.HandlerFunc {
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
		script, err := os.ReadFile(htmxFilePath)
		if err != nil {
			log.Error("could not load htmx from file", "error", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		var data indexData
		data.Script = string(script)
		if query := r.URL.Query().Get("search"); query != "" {
			videos, err := GetVideosHTML(db, query)
			if err != nil {
				log.Error("could not get videos", "error", err.Error())
			}
			data.Videos = *videos
		} else {
			data.Videos = "<h2>Search for something like em dash</h2>"
		}

		err = htmlTemplate.Execute(w, data)
		if err != nil {
			log.Error("could not write html to response", "error", err.Error())
		}
	})
}
