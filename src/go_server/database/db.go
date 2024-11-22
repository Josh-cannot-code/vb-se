package database

import (
	"context"
	"go_server/types"
)

type Repository interface {
	// TODO: transaction PutVideos
	PutVideo(ctx context.Context, vid *types.Video) error
	GetChannelIds(ctx context.Context) ([]string, error)
	GetVideoIds(ctx context.Context, channelId string) ([]string, error)
}
