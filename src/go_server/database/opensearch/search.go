package database

import (
	"encoding/json"
	"fmt"
	"go_server/models"
	"io"
	"os"
	"strings"

	"github.com/labstack/gommon/log"
)

type OpenSearchAccess struct {
	client *OpensearchHttpClient
}

func GetOpenSearchAccess() (*OpenSearchAccess, error) {
	client, err := getOpensearchHttpClient()
	if err != nil {
		return nil, err
	}

	if !client.Ping() {
		return nil, fmt.Errorf("failed to ping OpenSearch")
	}

	return &OpenSearchAccess{client: client}, nil
}

func getSearchBody(query string, sortField string, sortOrder string) string {
	modelId := os.Getenv("OPENSEARCH_MODEL_ID")
	return fmt.Sprintf(`{
		"_source": {
			"exclude": ["title_embedding", "transcript_embedding", "description_embedding"]
		},
		"sort": [
			{
				"%s": {
					"order": "%s"
				}
			}
		],
		"query": {
			"hybrid": {
				"queries": [
					{
						"multi_match": {
							"query": "%s",
							"fields": ["transcript", "title^3", "description^2"]
						}
					},
					{
						"neural": {
							"transcript_embedding": {
								"query_text": "%s",
								"model_id": "%s",
								"k": 5
							}
						}
					},
					{
						"neural": {
							"description_embedding": {
								"query_text": "%s",
								"model_id": "%s",
								"k": 5
							}
						}
					},
					{
						"neural": {
							"title_embedding": {
								"query_text": "%s",
								"model_id": "%s",
								"k": 5
							}
						}
					}
				]
			}
		},
		"search_pipeline": {
			"phase_results_processors": [
				{
					"normalization-processor": {
						"normalization": {
							"technique": "min_max"
						},
						"combination": {
							"technique": "arithmetic_mean",
							"parameters": {
							"weights": [0.3, 0.5, 0.1, 0.1]
							}
						},
						"ignore_failure": false
					}
				}
			]
		}
	}`, sortField, sortOrder, query, query, modelId, query, modelId, query, modelId)
}

func (c *OpenSearchAccess) SearchVideos(q string, sorting string) ([]*models.Video, error) {
	var sortField string
	var sortOrder string

	switch sorting {
	case "relevance":
		sortField = "_score"
		sortOrder = "desc"
	case "oldest":
		sortField = "upload_date"
		sortOrder = "asc"
	case "newest":
		sortField = "upload_date"
		sortOrder = "desc"
	default:
		sortField = "_score"
		sortOrder = "desc"
	}

	searchBody := getSearchBody(q, sortField, sortOrder)

	res, err := c.client.Get("/vb-se-videos/_search?size=50", strings.NewReader(searchBody))
	if err != nil {
		return nil, fmt.Errorf("request to opensearch failed: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode >= 400 {
		rawErr, _ := io.ReadAll(res.Body)
		if strings.Contains(string(rawErr), "Please deploy the model first") {
			log.Info("Model not deployed, attepmting deployment")
			_, err = c.client.Post("/_plugins/_ml/models/"+os.Getenv("OPENSEARCH_MODEL_ID")+"/_deploy", nil)
			if err != nil {
				return nil, fmt.Errorf("failed to make request to deploy model: %w", err)
			}

			if res.StatusCode >= 400 {
				rawErr, _ = io.ReadAll(res.Body)
				return nil, fmt.Errorf("error deploying model (status %d): %s", res.StatusCode, string(rawErr))
			}
		}
		return nil, fmt.Errorf("model was not deployed, re-deploy started")
	}

	// Parse the response
	var searchResponse struct {
		Hits struct {
			Hits []struct {
				Source models.Video `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(res.Body).Decode(&searchResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	videos := make([]*models.Video, 0, len(searchResponse.Hits.Hits))
	for _, hit := range searchResponse.Hits.Hits {
		videos = append(videos, &hit.Source)
	}

	return videos, nil
}
