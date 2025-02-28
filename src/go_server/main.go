package main

import (
	"context"
	"fmt"
	"go_server/database"
	"go_server/frontend"
	"go_server/youtube"
	"log/slog"
	"os"

	"net/http"

	"github.com/elastic/go-elasticsearch/v8"
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

	// Elastic
	log.Info("elastic host", "value", os.Getenv("ELASTIC_HOST"))

	var esClient *elasticsearch.TypedClient
	if os.Getenv("ENVIRONMENT") == "prod" {
		esClient, err = elasticsearch.NewTypedClient(elasticsearch.Config{
			Addresses: []string{
				os.Getenv("ELASTIC_HOST"),
			},
			APIKey: os.Getenv("ELASTIC_API_KEY"),
			CACert: []byte(os.Getenv("ELASTIC_CA_CERT")),
		})
	} else {
		esClient, err = elasticsearch.NewTypedClient(elasticsearch.Config{
			Addresses: []string{
				os.Getenv("ELASTIC_HOST"),
			},
			APIKey: os.Getenv("ELASTIC_API_KEY"),
		})
	}
	if err != nil {
		log.Error("failed to connect to elastic search", "message", err.Error())
	}

	res := esClient.Info
	// TODO: need res in loggable format
	log.Info("Elastic Info retrieved", "elastic_info", res)
	db := database.NewElasticConnection(esClient)

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
