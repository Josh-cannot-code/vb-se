package database

import (
	"go_server/models"
)

type Datasource interface {
	SearchVideos(query, index string) ([]*models.Video, error)
}
