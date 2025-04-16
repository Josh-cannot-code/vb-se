package frontend

import (
	"go_server/database"
	"go_server/types"
	"net/http"

	"github.com/labstack/echo/v4"
)

func GetVideos(db database.Repository, query string, sorting string) ([]*types.VCardData, error) {
	videos, err := db.SearchVideos(query, sorting)
	if err != nil {
		return nil, err
	}

	// Add description snippets
	for _, video := range videos {
		descLength := min(300, len(video.Video.Description))
		video.DescSnippet = video.Video.Description[:descLength]
	}

	return videos, nil
}

func GetSearchVideos(ctx echo.Context, db database.Repository) error {
	// Get query parameters
	query := ctx.QueryParam("search")
	sorting := ctx.QueryParam("sorting")

	// Get videos
	videos, err := GetVideos(db, query, sorting)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get videos")
	}

	// Handle empty results
	if len(videos) == 0 {
		return ctx.HTML(http.StatusOK, "<p>No results :(</p>")
	}

	// Return template response
	return ctx.Render(http.StatusOK, "video-card.html.tmpl", videos)
}
