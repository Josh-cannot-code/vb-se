package main

import (
	database "go_server/database/opensearch"
	"go_server/rest"

	"log/slog"
	"os"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

func main() {
	// Load environment variables
	err := godotenv.Load("../.env")
	if err != nil {
		slog.Warn(".env file not loaded")
	}

	// Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// OpenSearch
	log.Info("opensearch host: ", os.Getenv("OPENSEARCH_HOST"))

	db, err := database.GetOpenSearchAccess()
	if err != nil {
		log.Error("failed to create opensearch http client: ", err.Error())
		return
	}

	log.Info("OpenSearch connection established")

	// Routes
	e.GET("/", rest.HandleSearch(db))
	e.Static("/static", "./static")
	e.File("/favicon.ico", "./static/favicon.ico")

	// Start server
	e.Logger.Fatal(e.Start(":" + os.Getenv("PORT")))
}
