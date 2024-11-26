package frontend

import (
	"go_server/database"
	"html/template"
	"net/http"

	slogctx "github.com/veqryn/slog-context"
)

func SearchVideos(db database.Repository) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get logger
		ctx := slogctx.With(r.Context(), "function", "search videos")
		log := slogctx.FromCtx(ctx)
		// TODO: more info here
		query := r.URL.Query().Get("search")
		qlog := log.With("query", query)
		qlog.Info("searching videos")

		w.Header().Add("Content-Type", "text/html")

		const templateFilePath = "./frontend/templates/video-card.html.tmpl"
		htmlTemplate, err := template.ParseFiles(templateFilePath)
		if err != nil {
			qlog.Error("could not load html template from file", "error", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		videos, err := db.SearchVideos(query)
		if err != nil {
			qlog.Error("could not get videos from db", "error", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// TODO: videos can be length 0

		err = htmlTemplate.Execute(w, videos)
		if err != nil {
			qlog.Error("could not write html to response", "error", err.Error())
		}
	})
}
