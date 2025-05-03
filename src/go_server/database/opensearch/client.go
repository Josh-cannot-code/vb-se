package database

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"

	"github.com/labstack/gommon/log"
)

var (
	instance *OpensearchHttpClient
	once     sync.Once
)

type OpensearchHttpClient struct {
	client *http.Client
}

func getOpensearchHttpClient() (*OpensearchHttpClient, error) {
	once.Do(func() {
		var err error
		instance, err = newOpensearchHttpClient()
		if err != nil {
			log.Error("failed to create OpenSearch connection: ", err.Error())
		}
	})
	return instance, nil
}

func newOpensearchHttpClient() (*OpensearchHttpClient, error) {
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	return &OpensearchHttpClient{
		client: httpClient,
	}, nil
}

func (c *OpensearchHttpClient) Get(path string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s%s", os.Getenv("OPENSEARCH_HOST"), path), body)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(os.Getenv("OPENSEARCH_USERNAME"), os.Getenv("OPENSEARCH_PASSWORD"))
	req.Header.Set("Content-Type", "application/json")
	return c.client.Do(req)
}

func (c *OpensearchHttpClient) Post(path string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest("POST", fmt.Sprintf("%s%s", os.Getenv("OPENSEARCH_HOST"), path), body)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(os.Getenv("OPENSEARCH_USERNAME"), os.Getenv("OPENSEARCH_PASSWORD"))
	req.Header.Set("Content-Type", "application/json")
	return c.client.Do(req)
}

func (c *OpensearchHttpClient) Ping() bool {
	res, err := c.Get("/", nil)
	if err != nil {
		return false
	}
	defer res.Body.Close()
	return res.StatusCode == 200
}
