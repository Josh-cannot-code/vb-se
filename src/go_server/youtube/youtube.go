package youtube

import (
	"context"
	"encoding/json"
	"fmt"
	"go_server/database"
	"go_server/types"

	"log"
	"os/exec"

	"os"

	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

func GetVideoIds(ctx context.Context, channelId string) ([]string, error) {
	// Get api key
	ytApiKey := os.Getenv("YOUTUBE_API_KEY")

	service, err := youtube.NewService(ctx, option.WithAPIKey(ytApiKey))
	if err != nil {
		return nil, err
	}

	// Get channel videos playlist id
	part := []string{"contentDetails"}
	channelCall := service.Channels.List(part)
	channelCall.Id(channelId)
	channelResp, err := channelCall.Do()
	if err != nil {
		return nil, err
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
			return nil, err
		}

		// Save video ids
		for _, item := range playlistResp.Items {
			videoIds = append(videoIds, item.ContentDetails.VideoId)
		}

		nextPageToken = playlistResp.NextPageToken
		if nextPageToken == "" {
			break
		}
	}

	return videoIds, nil
}

func GetVideo(videoId string) (*types.Video, error) {
	path, err := exec.LookPath("python3")
	if err != nil {
		return nil, err
	}

	// TODO: need better errors here so we know when to backoff video gets
	out, err := exec.Command(path, "download_video.py", videoId).Output()
	if err != nil {
		return nil, err
	}

	var vid types.Video
	err = json.Unmarshal(out, &vid)
	if err != nil {
		return nil, err
	}

	return &vid, nil
}

// Get all elements in a that do not appear in b
// len(a) >= len(b)
func inAnotInB(a []string, b []string) []string {
	diff := make(map[string]bool, len(a))

	for _, e := range a {
		diff[e] = true
	}

	for _, e := range b {
		diff[e] = false
	}

	var result []string
	for k, v := range diff {
		if v {
			result = append(result, k)
		}
	}
	return result
}

func GetVideoTranscripts(videoIds []string) map[string]string {
	if len(videoIds) == 0 {
		return make(map[string]string)
	}
	path, err := exec.LookPath("python3")
	if err != nil {
		log.Fatal(err)
	}

	vidsArg := videoIds[0]
	for _, vId := range videoIds[1:] {
		vidsArg += "," + vId
	}

	out, err := exec.Command(path, "get_transcripts.py", vidsArg).Output()
	if err != nil {
		log.Fatal(err)
	}

	transcriptMap := make(map[string]string)
	err = json.Unmarshal(out, &transcriptMap)
	if err != nil {
		log.Fatal(err)
	}

	return transcriptMap
}

func RefreshVideos(ctx context.Context, db database.Repository) error {
	channelIds, err := db.GetChannelIds(ctx)
	if err != nil {
		return err
	}

	for _, channelId := range channelIds {
		// TODO: replace with proper logger
		fmt.Printf("getting videos for channel: %s\n", channelId)

		storedVideoIds, err := db.GetVideoIds(ctx, channelId)
		if err != nil {
			return err
		}

		// TODO: remove deleted videos from db too?
		allVideoIds, err := GetVideoIds(ctx, channelId)
		if err != nil {
			return err
		}

		videosToGet := inAnotInB(allVideoIds, storedVideoIds)
		// TODO: this can be more efficient with transactions, temp solution
		for _, vId := range videosToGet {
			fmt.Printf("getting video: %s\n", vId)
			video, err := GetVideo(vId)
			if err != nil {
				return err
			}
			fmt.Printf("got video: %s\n", vId)
			err = db.PutVideo(ctx, video)
			if err != nil {
				return err
			}
			fmt.Printf("video: %s put in database\n", vId)
		}
		// TODO: logging here with channel ids to know when we are finished with one
		// TODO: transcripts and text strings for videos that don't have them
	}
	return nil
}
