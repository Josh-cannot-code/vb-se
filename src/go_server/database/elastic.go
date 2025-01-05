package database

import (
	"context"
	"encoding/json"
	"fmt"
	internalTypes "go_server/types"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/typedapi/esql/query"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/sortorder"
)

type ElasticConnection struct {
	es *elasticsearch.TypedClient
}

func NewElasticConnection(es *elasticsearch.TypedClient) *ElasticConnection {
	return &ElasticConnection{
		es: es,
	}
}

func (c ElasticConnection) PutVideo(ctx context.Context, video *internalTypes.Video) error {
	// TODO: maybe need elastic video type
	_, err := c.es.Index("vb-se-videos").Request(video).Do(ctx)
	return err
}

type Channel struct {
	ChannelName string `json:"channel_name"`
	ChannelID   string `json:"channel_id"`
}

func (c ElasticConnection) CreateIfNoIndices(ctx context.Context) error {
	// Create Ingest Pipeline
	dateProcessor := &types.ProcessorContainer{
		Set: &types.SetProcessor{
			Field: "created_at",
			Value: json.RawMessage(`"{{_ingest.timestamp}}"`),
		},
	}
	_, err := c.es.Ingest.PutPipeline("ingest_with_dates").Processors(*dateProcessor).Do(ctx)
	if err != nil {
		return err
	}

	if exists, err := c.es.Indices.Exists("vb-se-videos").Do(ctx); err == nil && !exists {

		mappings := types.NewTypeMapping()
		mappings.Properties["semantic_text"] = &types.SemanticTextProperty{
			InferenceId: ".elser-2-elasticsearch",
			Type:        "semantic_text",
		}
		mappings.Properties["transcript"] = &types.TextProperty{
			CopyTo: []string{"semantic_text"},
		}
		mappings.Properties["title"] = &types.TextProperty{
			CopyTo: []string{"semantic_text"},
		}
		mappings.Properties["description"] = &types.TextProperty{
			CopyTo: []string{"semantic_text"},
		}
		mappings.Properties["video_id"] = types.NewTextProperty()
		mappings.Properties["channel_id"] = types.NewTextProperty()
		mappings.Properties["thumbnail"] = types.NewTextProperty()
		mappings.Properties["url"] = types.NewTextProperty()
		mappings.Properties["upload_date"] = types.NewTextProperty()
		mappings.Properties["channel_name"] = types.NewTextProperty()
		mappings.Properties["created_at"] = types.NewDateProperty()

		settings := types.NewIndexSettings()
		ingestPipelineName := "ingest_with_dates"
		settings.DefaultPipeline = &ingestPipelineName

		_, err := c.es.Indices.Create("vb-se-videos").Settings(settings).Mappings(mappings).Do(ctx)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	if exists, err := c.es.Indices.Exists("vb-se-channels").Do(ctx); err == nil && !exists {

		mappings := types.NewTypeMapping()
		mappings.Properties["channel_id"] = types.NewTextProperty()
		mappings.Properties["channel_name"] = types.NewTextProperty()
		mappings.Properties["created_at"] = types.NewDateProperty()

		settings := types.NewIndexSettings()
		ingestPipelineName := "ingest_with_dates"
		settings.DefaultPipeline = &ingestPipelineName

		_, err := c.es.Indices.Create("vb-se-channels").Settings(settings).Mappings(mappings).Do(ctx)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	return nil
}

func (c ElasticConnection) GetNoTranscriptVideoIds(ctx context.Context) ([]string, error) {
	return nil, nil
}
func (c ElasticConnection) UpdateVideoTranscripts(ctx context.Context, transcriptMap map[string]string) error {
	return nil
}
func (c ElasticConnection) UpdateVideoTextData(ctx context.Context) error { return nil }

// TODO: these should not use append, static size instead
func (c ElasticConnection) GetChannelIds(ctx context.Context) ([]string, error) {
	queryChannels := `from vb-se-channels`
	qry := c.es.Esql.Query().Query(queryChannels)
	channels, err := query.Helper[Channel](ctx, qry)
	if err != nil {
		return nil, err
	}

	var channelIds []string
	for _, channel := range channels {
		channelIds = append(channelIds, channel.ChannelID)
	}
	return channelIds, nil
}

func (c ElasticConnection) GetVideoIds(ctx context.Context, channelId string) ([]string, error) {
	queryVideos := `from vb-se-videos`
	qry := c.es.Esql.Query().Query(queryVideos)
	videos, err := query.Helper[internalTypes.Video](ctx, qry)
	if err != nil {
		return nil, err
	}

	var videoIds []string
	for _, video := range videos {
		videoIds = append(videoIds, video.VideoID)
	}
	return videoIds, nil
}

func (c ElasticConnection) SearchVideos(q string, sorting string) ([]internalTypes.VCardData, error) {
	// TODO construct sort here and pass in the sort dir
	var sortOpts map[string]types.FieldSort
	switch sorting {
	case "relevece":
		sortOpts = map[string]types.FieldSort{
			"_score": {
				Order: &sortorder.Desc,
			},
		}
	case "oldest":
		sortOpts = map[string]types.FieldSort{
			"upload_date": {
				Order: &sortorder.Asc,
			},
		}
	case "newest":
		sortOpts = map[string]types.FieldSort{
			"upload_date": {
				Order: &sortorder.Desc,
			},
		}
	}
	fmt.Print(sortOpts)
	rankWindowSize := 30
	resp, err := c.es.Search().Index("vb-se-videos").Retriever(&types.RetrieverContainer{
		// TODO: boosting
		// TODO: the sorting here HAS to be wrong since we are doing it pre rrf
		Rrf: &types.RRFRetriever{
			RankWindowSize: &rankWindowSize,
			Retrievers: []types.RetrieverContainer{
				{
					Standard: &types.StandardRetriever{
						Query: &types.Query{
							MultiMatch: &types.MultiMatchQuery{
								Fields: []string{
									"transcript",
									"description",
									"title",
								},
								Query: q,
							},
						},
					},
				},
				{
					Standard: &types.StandardRetriever{
						Query: &types.Query{
							Semantic: &types.SemanticQuery{
								Field: "semantic_text",
								Query: q,
							},
						},
					},
				},
			},
		},
	}).From(0).Size(rankWindowSize).
		Do(context.Background())
	if err != nil {
		return nil, err
	}

	// TODO: initialize with the correct length
	var vCards []internalTypes.VCardData
	for _, hit := range resp.Hits.Hits {
		var curVideo internalTypes.Video
		err := json.Unmarshal(hit.Source_, &curVideo)
		if err != nil {
			return nil, err
		}

		var curVCardData internalTypes.VCardData
		curVCardData.Video = curVideo
		vCards = append(vCards, curVCardData)
	}

	return vCards, nil
}
