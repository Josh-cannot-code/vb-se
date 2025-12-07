package rest

import (
	"go_server/components"
	database "go_server/database"

	"go_server/models"
	"net/http"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
)

func HandleSearch(db database.Datasource, index string) echo.HandlerFunc {
	return func(c echo.Context) error {
		query := c.QueryParam("search")

		var videos []*models.Video
		if query != "" {
			var err error
			videos, err = db.SearchVideos(query, index)
			if err != nil {
				c.Logger().Error("failed to search videos: ", err.Error())
				return echo.NewHTTPError(http.StatusInternalServerError, "Failed to search videos")
			}
		}

		return render(c, http.StatusOK, components.Index(videos))
	}
}

func render(ctx echo.Context, statusCode int, t templ.Component) error {
	ctx.Response().Writer.WriteHeader(statusCode)
	ctx.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTML)
	return t.Render(ctx.Request().Context(), ctx.Response().Writer)
}
