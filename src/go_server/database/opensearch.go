package database

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	internalTypes "go_server/types"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/opensearch-project/opensearch-go"
	"github.com/opensearch-project/opensearch-go/opensearchapi"
	slogctx "github.com/veqryn/slog-context"
)

type OpenSearchConnection struct {
	osc *opensearch.Client
}

func NewOpenSearchConnection() (*OpenSearchConnection, error) {
	// Create OpenSearch client with authentication
	cfg := opensearch.Config{
		Addresses: []string{os.Getenv("OPENSEARCH_HOST")},
		Username:  os.Getenv("OPENSEARCH_USERNAME"),
		Password:  os.Getenv("OPENSEARCH_PASSWORD"),
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	client, err := opensearch.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create OpenSearch client: %w", err)
	}

	resp, err := client.Ping()
	if err != nil {
		return nil, fmt.Errorf("failed to ping OpenSearch: %w", err)
	}
	defer resp.Body.Close()

	// Read and print response body
	_, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read ping response: %w", err)
	}

	return &OpenSearchConnection{
		osc: client,
	}, nil
}

func (c OpenSearchConnection) PutVideo(ctx context.Context, video *internalTypes.Video) error {
	// OpenSearch does not allow empty strings for embedding
	if video.Transcript == "" {
		video.Transcript = "No transcript available"
	}

	if video.Description == "" {
		video.Description = "No description available"
	}

	body, err := json.Marshal(video)
	if err != nil {
		return fmt.Errorf("failed to marshal video: %w", err)
	}

	req := opensearchapi.IndexRequest{
		Index:      "vb-se-videos",
		Body:       strings.NewReader(string(body)),
		DocumentID: video.VideoID,
	}

	res, err := req.Do(ctx, c.osc)
	if err != nil {
		return fmt.Errorf("failed to index video: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error indexing video: %s", res.String())
	}

	return nil
}

func (c OpenSearchConnection) GetNoTranscriptVideoIds(ctx context.Context) ([]*string, error) {
	// TODO: Implement OpenSearch version
	return nil, nil
}

func (c OpenSearchConnection) UpdateVideoTranscripts(ctx context.Context, transcriptMap map[string]string) error {
	// TODO: Implement OpenSearch version
	return nil
}

func (c OpenSearchConnection) UpdateVideoTextData(ctx context.Context) error {
	// TODO: Implement OpenSearch version
	return nil
}

func (c OpenSearchConnection) GetChannelIds(ctx context.Context) ([]*string, error) {
	// Create search request to get all channels
	searchBody := strings.NewReader(`{
		"query": {
			"match_all": {}
		},
		"_source": ["channel_id"]
	}`)

	searchReq := opensearchapi.SearchRequest{
		Index: []string{"vb-se-channels"},
		Body:  searchBody,
	}

	res, err := searchReq.Do(ctx, c.osc)
	if err != nil {
		return nil, fmt.Errorf("failed to search channels: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("error searching channels: %s", res.String())
	}

	// Parse the response
	var searchResponse struct {
		Hits struct {
			Hits []struct {
				Source struct {
					ChannelID string `json:"channel_id"`
				} `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(res.Body).Decode(&searchResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Extract channel IDs
	channelIds := make([]*string, 0, len(searchResponse.Hits.Hits))
	for _, hit := range searchResponse.Hits.Hits {
		channelIds = append(channelIds, &hit.Source.ChannelID)
	}

	return channelIds, nil
}

func (c OpenSearchConnection) GetVideoIds(ctx context.Context, channelId string) ([]*string, error) {
	// Create search query for videos by channel ID
	searchBody := strings.NewReader(fmt.Sprintf(`{
		"query": {
			"term": {
				"channel_id": {
					"value": "%s"
				}
			}
		},
		"_source": ["video_id"]
	}`, channelId))

	searchSize := 10000

	// Create search request
	searchReq := opensearchapi.SearchRequest{
		Index: []string{"vb-se-videos"},
		Body:  searchBody,
		Size:  &searchSize,
	}

	// Execute search
	res, err := searchReq.Do(ctx, c.osc)
	if err != nil {
		return nil, fmt.Errorf("failed to execute search: %w", err)
	}
	defer res.Body.Close()

	// Check for errors in response
	if res.IsError() {
		return nil, fmt.Errorf("search request failed: %s", res.String())
	}

	// Parse response
	var searchResponse struct {
		Hits struct {
			Hits []struct {
				Source struct {
					VideoID string `json:"video_id"`
				} `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(res.Body).Decode(&searchResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Extract video IDs
	videoIds := make([]*string, 0, len(searchResponse.Hits.Hits))
	for _, hit := range searchResponse.Hits.Hits {
		videoIds = append(videoIds, &hit.Source.VideoID)
	}

	// Get logger
	ctx = slogctx.With(ctx, "function", "GetVideoIds")
	log := slogctx.FromCtx(ctx)

	log.Info("GetVideoIds", "channelId", channelId, "numVideosInDb", len(videoIds))

	return videoIds, nil
}

func (c OpenSearchConnection) SearchVideos(q string, sorting string) ([]*internalTypes.Video, error) {
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

	modelId := os.Getenv("OPENSEARCH_MODEL_ID")

	// Construct the search query
	searchBody := strings.NewReader(fmt.Sprintf(`{
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
	}`, sortField, sortOrder, q, q, modelId, q, modelId, q, modelId))

	// Create HTTP request
	searchUrl := fmt.Sprintf("%s/vb-se-videos/_search?size=50", os.Getenv("OPENSEARCH_HOST"))
	req, err := http.NewRequest("GET", searchUrl, searchBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers and auth
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(os.Getenv("OPENSEARCH_USERNAME"), os.Getenv("OPENSEARCH_PASSWORD"))

	// Create insecure http client to make search request
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	// Execute request
	res, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request to opensearch failed: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode >= 400 {
		rawErr, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("error searching videos (status %d): %s", res.StatusCode, string(rawErr))
	}

	// Parse the response
	var searchResponse struct {
		Hits struct {
			Hits []struct {
				Source internalTypes.Video `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(res.Body).Decode(&searchResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	videos := make([]*internalTypes.Video, 0, len(searchResponse.Hits.Hits))
	for _, hit := range searchResponse.Hits.Hits {
		videos = append(videos, &hit.Source)
	}

	return videos, nil
}
