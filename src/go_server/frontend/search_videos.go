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
		ctx := slogctx.With(r.Context(), "function", "index")
		log := slogctx.FromCtx(ctx)

		w.Header().Add("Content-Type", "text/html")

		const templateFilePath = "./frontend/templates/video-card.html.tmpl"
		htmlTemplate, err := template.ParseFiles(templateFilePath)
		if err != nil {
			log.Error("could not load html template from file", "error", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		videos, err := db.GetVideos()
		if err != nil {
			log.Error("could not get videos from db", "error", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		err = htmlTemplate.Execute(w, videos[0])
		if err != nil {
			log.Error("could not write html to response", "error", err.Error())
		}
	})
}
