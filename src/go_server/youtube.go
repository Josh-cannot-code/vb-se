package main

import (
	"context"
	"fmt"
	"log"

	"os"

	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

// TODO: handle errors
func getVideoIds(ctx context.Context, channelId string) []string {
	// Get api key
	ytApiKey := os.Getenv("YOUTUBE_API_KEY")

	service, err := youtube.NewService(ctx, option.WithAPIKey(ytApiKey))
	if err != nil {
		log.Fatal(err)
	}

	// Get channel videos playlist id
	part := []string{"contentDetails"}
	channelCall := service.Channels.List(part)
	channelCall.Id(channelId)
	channelResp, err := channelCall.Do()
	if err != nil {
		log.Fatal(err)
	}
	playlstId := channelResp.Items[0].ContentDetails.RelatedPlaylists.Uploads

	// Get all videos from channel
	var videoIds []string
	nextPageToken := ""
	for {
		playlistCall := service.PlaylistItems.List(part)
		playlistCall.PlaylistId(playlstId)
		playlistCall.MaxResults(50)
		if nextPageToken != "" {
			playlistCall.PageToken(nextPageToken)
		}

		playlistResp, err := playlistCall.Do()
		if err != nil {
			log.Fatal(err)
		}

		// Save video ids
		for _, item := range playlistResp.Items {
			videoIds = append(videoIds, item.Id)
		}

		nextPageToken = playlistResp.NextPageToken
		fmt.Println(nextPageToken)
		if nextPageToken == "" {
			break
		}
	}

	return videoIds
}
