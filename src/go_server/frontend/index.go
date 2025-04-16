package frontend

//import (
//	"go_server/database"
//	"net/http"
//	"os"
//	"text/template"
//
//	"github.com/labstack/echo/v4"
//	slogctx "github.com/veqryn/slog-context"
//)
//
//type indexData struct {
//	Videos string
//	Script string
//}
//
//const indexFilePath = "./frontend/templates/index.html.tmpl"
//
//func GetIndex(ctx echo.Context, db database.Repository) error {
//	// Read htmx script
//	const htmxFilePath = "./frontend/templates/htmx.min.js"
//	script, err := os.ReadFile(htmxFilePath)
//	if err != nil {
//		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to load required scripts")
//	}
//
//	// Prepare template data
//	data := indexData{
//		Script: string(script),
//	}
//
//	// Handle search query if present
//	if query := ctx.QueryParam("search"); query != "" {
//		sorting := ctx.QueryParam("sorting")
//		videos, err := GetVideos(db, query, sorting)
//		if err != nil {
//			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get videos")
//		}
//		if len(videos) == 0 {
//			data.Videos = "<p>No results :(</p>"
//		} else {
//			//data.Videos = videos
//			return ctx.Render(http.StatusOK, "video-card.html.tmpl", videos)
//		}
//	}
//
//	// Render template
//	return ctx.Render(http.StatusOK, "index.html.tmpl", data)
//}
//
//func Index(db database.Repository) http.HandlerFunc {
//	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//		// Get logger
//		ctx := slogctx.With(r.Context(), "function", "index")
//		log := slogctx.FromCtx(ctx)
//
//		w.Header().Add("Content-Type", "text/html")
//
//		htmlTemplate, err := template.ParseFiles(indexFilePath)
//		if err != nil {
//			log.Error("could not load html template from file", "error", err.Error())
//			w.WriteHeader(http.StatusInternalServerError)
//			return
//		}
//
//		const htmxFilePath = "./frontend/templates/htmx.min.js"
//		script, err := os.ReadFile(htmxFilePath)
//		if err != nil {
//			log.Error("could not load htmx from file", "error", err.Error())
//			w.WriteHeader(http.StatusInternalServerError)
//			return
//		}
//
//		var data indexData
//		data.Script = string(script)
//		// Get query
//		query := r.URL.Query().Get("search")
//		if query != "" {
//			sorting := r.URL.Query().Get("sorting")
//			qlog := log.With("query", query, "sorting", sorting)
//			qlog.Info("searching videos")
//
//			w.Header().Add("Content-Type", "text/html")
//
//			videos, err := GetVideos(db, query, sorting)
//			if err != nil {
//				qlog.Error("could not get videos", "error", err.Error())
//				w.WriteHeader(http.StatusInternalServerError)
//				return
//			}
//			if len(videos) == 0 {
//				data.Videos = "<p>No results :(</p>"
//			} else {
//				err = htmlTemplate.Execute(w, videos)
//				if err != nil {
//					qlog.Error("could not execute template", "error", err.Error())
//					w.WriteHeader(http.StatusInternalServerError)
//					return
//				}
//				return
//			}
//		} else {
//			data.Videos = "<h2>Search for something like em dash</h2>"
//		}
//
//		err = htmlTemplate.Execute(w, data)
//	})
//}
//
