package frontend

import (
	"go_server/database"
	"net/http"
	"strings"
	"text/template"

	slogctx "github.com/veqryn/slog-context"
)

func GetVideosHTML(db database.Repository, query string) (*string, error) {
	const templateFilePath = "./frontend/templates/video-card.html.tmpl"
	htmlTemplate, err := template.ParseFiles(templateFilePath)
	if err != nil {
		return nil, err
	}

	videos, err := db.SearchVideos(query)
	if err != nil {
		return nil, err
	}

	if len(videos) == 0 {
		s := "<p>No results :(</p>"
		return &s, nil
	} else {
		var vsb strings.Builder
		err = htmlTemplate.Execute(&vsb, videos)
		if err != nil {
			return nil, err
		}
		vs := vsb.String()
		return &vs, nil
	}

}

func SearchVideos(db database.Repository) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get logger
		ctx := slogctx.With(r.Context(), "function", "search videos")
		log := slogctx.FromCtx(ctx)
		// Get query
		query := r.URL.Query().Get("search")
		qlog := log.With("query", query)
		qlog.Info("searching videos")

		w.Header().Add("Content-Type", "text/html")

		videoString, err := GetVideosHTML(db, query)
		if err != nil {
			qlog.Error("could not get videos", "error", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
		}
		_, err = w.Write([]byte(*videoString))
		if err != nil {
			qlog.Error("could not write videos to response", "error", err.Error())
		}
	})
}
