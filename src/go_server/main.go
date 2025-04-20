package main

import (
	"go_server/components"
	"go_server/database"

	"go_server/types"
	"go_server/youtube"
	"log/slog"
	"net/http"
	"os"

	"github.com/a-h/templ"
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

	// Initialize logger
	//defaultAttrs := []slog.Attr{
	//	slog.String("service", "vb-be"),
	//	slog.String("environment", os.Getenv("ENVIRONMENT")),
	//}

	// Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	//baseHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{AddSource: true}).WithAttrs(defaultAttrs)
	//customHandler := slogctx.NewHandler(baseHandler, nil)
	//slog.SetDefault(slog.New(customHandler))

	//ctx := slogctx.NewCtx(context.Background(), slog.Default())
	//log := slogctx.FromCtx(ctx)

	// OpenSearch
	log.Info("opensearch host: ", os.Getenv("OPENSEARCH_HOST"))

	db, err := database.NewOpenSearchConnection()
	if err != nil {
		log.Error("failed to connect to opensearch: ", err.Error())
		return
	}
	log.Info("OpenSearch connection established")

	// Handler declarations
	refreshHandler := youtube.RefreshVideos(db)

	// Routes
	e.Static("/static", "./static")

	e.GET("/refresh", func(c echo.Context) error {
		refreshHandler(c.Response(), c.Request())
		return nil
	})
	e.GET("/", func(c echo.Context) error {
		query := c.QueryParam("search")

		var videos []*types.Video
		err := error(nil)
		if query != "" {
			videos, err = db.SearchVideos(query, "relevance")
			if err != nil {
				e.Logger.Error("failed to search videos: ", err.Error())
				return echo.NewHTTPError(http.StatusInternalServerError, "Failed to search videos")
			}
		}

		return render(c, http.StatusOK, components.Index(videos))
	})

	// Start server
	e.Logger.Fatal(e.Start(":" + os.Getenv("PORT")))
}

func render(ctx echo.Context, statusCode int, t templ.Component) error {
	ctx.Response().Writer.WriteHeader(statusCode)
	ctx.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTML)
	return t.Render(ctx.Request().Context(), ctx.Response().Writer)
}
