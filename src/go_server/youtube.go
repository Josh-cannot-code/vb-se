package main

import (
	"context"
	"encoding/json"
	"fmt"

	//	"io"
	"log"
	"os/exec"

	"os"

	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

type video struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Thumbnail   string `json:"thumbnail"`
	ChannelID   string `json:"channel_id"`
	Description string `json:"description"`
	UploadDate  string `json:"upload_date"` // TODO: proper datetime
	URL         string `json:"url"`
	ChannelName string `json:"channel_name"`
}

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
			videoIds = append(videoIds, item.ContentDetails.VideoId)
		}

		nextPageToken = playlistResp.NextPageToken
		fmt.Println(nextPageToken)
		if nextPageToken == "" {
			break
		}
	}

	return videoIds
}

func getVideo(videoId string) video {
	path, err := exec.LookPath("python3")
	if err != nil {
		log.Fatal(err)
	}
	out, err := exec.Command(path, "download_video.py").Output() // TODO: + videoId)
	if err != nil {
		log.Fatal(err)
	}

	var vid video
	err = json.Unmarshal(out, &vid)
	if err != nil {
		log.Fatal(err)
	}

	return vid
}
