package database

import (
	"go_server/models"
)

type Datasource interface {
	SearchVideos(query string, sorting string) ([]*models.Video, error)
}
