package main

import (
	"context"
	"database/sql"
	"fmt"
	"go_server/database"
	"go_server/frontend"
	"go_server/youtube"
	"log/slog"
	"os"

	"net/http"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
	slogctx "github.com/veqryn/slog-context"
)

func LoggerMiddleware(next http.Handler, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add slogger to the context
		ctx := slogctx.NewCtx(r.Context(), logger)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func main() {
	// Load environment variables
	err := godotenv.Load(".env")
	if err != nil {
		panic("Could not load env")
	}

	// Initialize logger
	// TODO: add info about location and stuff here
	defaultAttrs := []slog.Attr{
		slog.String("service", "vb-be"),
		slog.String("environment", "dev"), // TODO: dev prod envs
	}

	baseHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{AddSource: true}).WithAttrs(defaultAttrs)
	customHandler := slogctx.NewHandler(baseHandler, nil)
	slog.SetDefault(slog.New(customHandler))

	ctx := slogctx.NewCtx(context.Background(), slog.Default())
	log := slogctx.FromCtx(ctx)

	sqlDb, err := sql.Open("sqlite3", "../db/vb-se.db")
	if err != nil {
		log.Error("could not connect to db", "message", err.Error())
		return
	}
	if err = sqlDb.Ping(); err != nil {
		log.Error("not connected to db", "message", err.Error())
		return
	}
	defer sqlDb.Close()
	db := database.NewSqLiteConnection(sqlDb)

	// Handler declarations
	refreshHandler := youtube.RefreshVideos(db)
	indexHandler := frontend.Index()
	searchVideosHandler := frontend.SearchVideos(db)

	r := mux.NewRouter()
	// Register handlers
	r.Handle("/refresh", refreshHandler).Methods("GET")
	r.Handle("/", indexHandler).Methods("GET")
	r.Handle("/videos", searchVideosHandler).Methods("GET")

	fmt.Println("Listening on http://localhost:3001")
	err = http.ListenAndServe(":3001", LoggerMiddleware(r, log))
	if err != nil {
		panic(err)
	}
}
