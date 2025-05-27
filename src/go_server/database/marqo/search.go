package database

import (
	"fmt"
	"go_server/models"
)

const vbSeIndex = "vb-se-videos"

type MarqoAccess struct {
	client *MarqoClient
}

func GetMarqoAccess(host string) (*MarqoAccess, error) {
	client, err := getMarqoClient(host)
	if err != nil {
		return nil, err
	}

	if !client.Ping() {
		return nil, fmt.Errorf("failed to ping Marqo")
	}

	return &MarqoAccess{client: client}, nil
}

func (c *MarqoAccess) SearchVideos(query string) ([]*models.Video, error) {
	return c.client.Search(query, vbSeIndex)
}
