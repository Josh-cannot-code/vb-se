package database

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"go_server/models"
	"net/http"
	"sync"

	"github.com/labstack/gommon/log"
)

var (
	instance *MarqoClient
	once     sync.Once
)

type MarqoClient struct {
	client *http.Client
	host   string
}

func newMarqoClient(host string) (*MarqoClient, error) {
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	return &MarqoClient{
		client: httpClient,
		host:   host,
	}, nil
}

func getMarqoClient(host string) (*MarqoClient, error) {
	once.Do(func() {
		var err error
		instance, err = newMarqoClient(host)
		if err != nil {
			log.Error("failed to create Marqo connection: ", err.Error())
		}
	})
	return instance, nil
}

func (c *MarqoClient) Ping() bool {
	res, err := c.client.Get(c.host)
	if err != nil {
		log.Error("failed to ping Marqo: ", err.Error())
		return false
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		log.Error("failed to ping Marqo: ", res.StatusCode)
		return false
	}

	return true
}

func (c *MarqoClient) Search(query string, index string) ([]*models.Video, error) {
	// TODO: sanitize query in access layer?
	reqBody := map[string]string{"q": query}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/%s", c.host, fmt.Sprintf("indexes/%s/search", index)), bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}

	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var searchResponse struct {
		Hits []*models.Video `json:"hits"`
	}
	if err := json.NewDecoder(res.Body).Decode(&searchResponse); err != nil {
		return nil, err
	}

	return searchResponse.Hits, nil
}
