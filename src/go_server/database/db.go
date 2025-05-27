package database

import (
	"go_server/models"
)

type Datasource interface {
	SearchVideos(query string) ([]*models.Video, error)
}
