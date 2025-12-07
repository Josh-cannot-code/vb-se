package database

import (
	"fmt"
	"go_server/models"
	"time"
)

const maxRetries = 10

type MarqoAccess struct {
	client *MarqoClient
}

func GetMarqoAccess(host string) (*MarqoAccess, error) {
	client, err := getMarqoClient(host)
	if err != nil {
		return nil, err
	}

	retryCount := 0
	sleepDuration := 2 * time.Second
	for !client.Ping() && retryCount < maxRetries {
		fmt.Println("Waiting for Marqo to be available...")
		time.Sleep(sleepDuration)
		sleepDuration *= 2
	}

	if !client.Ping() {
		return nil, fmt.Errorf("failed to ping Marqo")
	}

	return &MarqoAccess{client: client}, nil
}

func (c *MarqoAccess) SearchVideos(query, index string) ([]*models.Video, error) {
	return c.client.Search(query, index)
}
