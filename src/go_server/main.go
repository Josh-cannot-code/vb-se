package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"go_server/database"
	"go_server/frontend"
	"go_server/youtube"
	"log/slog"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/opensearch-project/opensearch-go"
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
	err := godotenv.Load("../.env")
	if err != nil {
		slog.Warn(".env file not loaded")
	}

	// Initialize logger
	// TODO: add info about location and stuff here
	defaultAttrs := []slog.Attr{
		slog.String("service", "vb-be"),
		slog.String("environment", os.Getenv("ENVIRONMENT")), // TODO: dev prod envs
	}

	baseHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{AddSource: true}).WithAttrs(defaultAttrs)
	customHandler := slogctx.NewHandler(baseHandler, nil)
	slog.SetDefault(slog.New(customHandler))

	ctx := slogctx.NewCtx(context.Background(), slog.Default())
	log := slogctx.FromCtx(ctx)

	// OpenSearch
	log.Info("opensearch host", "value", os.Getenv("OPENSEARCH_HOST"))

	var osConfig opensearch.Config
	if os.Getenv("ENVIRONMENT") == "prod" {
		osConfig = opensearch.Config{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: false,
				},
			},
			Addresses: []string{
				os.Getenv("OPENSEARCH_HOST"),
			},
			Username: os.Getenv("OPENSEARCH_USERNAME"),
			Password: os.Getenv("OPENSEARCH_PASSWORD"),
		}
	} else {
		osConfig = opensearch.Config{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
			Addresses: []string{
				os.Getenv("OPENSEARCH_HOST"),
			},
			Username: os.Getenv("OPENSEARCH_USERNAME"),
			Password: os.Getenv("OPENSEARCH_PASSWORD"),
		}
	}

	osClient, err := opensearch.NewClient(osConfig)
	if err != nil {
		log.Error("failed to connect to opensearch", "message", err.Error())
	}

	db := database.NewOpenSearchConnection(osClient)
	log.Info("OpenSearch connection established")

	// Make sure elastic indices have been created
	err = db.CreateIfNoIndices(ctx)
	if err != nil {
		log.Error("could not refresh indices", "message", err.Error())
	}

	// Handler declarations
	refreshHandler := youtube.RefreshVideos(db)
	indexHandler := frontend.Index(db)
	searchVideosHandler := frontend.SearchVideos(db)

	r := mux.NewRouter()
	// Register handlers
	r.Handle("/refresh", refreshHandler).Methods("GET")
	r.Handle("/", indexHandler).Methods("GET")
	r.Handle("/videos", searchVideosHandler).Methods("GET")

	port := os.Getenv("PORT")
	fmt.Println("Listening on http://localhost:" + port)
	err = http.ListenAndServe(":"+port, LoggerMiddleware(r, log))
	if err != nil {
		panic(err)
	}
}
