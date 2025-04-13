package database

import (
	"context"
	"go_server/types"
)

type Repository interface {
	// TODO: transaction PutVideos
	PutVideo(ctx context.Context, vid *types.Video) error
	GetChannelIds(ctx context.Context) ([]*string, error)
	GetVideoIds(ctx context.Context, channelId string) ([]*string, error)
	GetNoTranscriptVideoIds(ctx context.Context) ([]*string, error)
	UpdateVideoTranscripts(ctx context.Context, transcriptMap map[string]string) error
	UpdateVideoTextData(ctx context.Context) error
	SearchVideos(query string, sorting string) ([]*types.VCardData, error)
}
