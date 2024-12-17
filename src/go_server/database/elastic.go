package database

import (
	"context"
	"encoding/json"
	internalTypes "go_server/types"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/typedapi/esql/query"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
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

func (c ElasticConnection) SearchVideos(q string) ([]internalTypes.VCardData, error) {
	resp, err := c.es.Search().Index("vb-se-videos").Retriever(&types.RetrieverContainer{
		// TODO: boosting
		Rrf: &types.RRFRetriever{
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
	}).Do(context.Background())
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
