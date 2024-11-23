package youtube

import (
	"context"
	"encoding/json"
	"go_server/database"
	"go_server/types"
	"net/http"

	"log"
	"os/exec"

	"os"

	slogctx "github.com/veqryn/slog-context"
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
	out, err := exec.Command(path, "python_scripts/get_video.py", videoId).Output()
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

	out, err := exec.Command(path, "python_scripts/get_transcripts.py", vidsArg).Output()
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

func RefreshVideos(db database.Repository) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get logger
		ctx := slogctx.With(r.Context(), "function", "refreshVideos")
		log := slogctx.FromCtx(ctx)

		channelIds, err := db.GetChannelIds(ctx)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log = log.With("error", err.Error())
			log.Error("could not get channel ids")
			return
		}

		for _, channelId := range channelIds {
			clog := log.With("channel_id", channelId)
			clog.Info("getting videos for channel")

			storedVideoIds, err := db.GetVideoIds(ctx, channelId)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				clog = log.With("error", err.Error())
				clog.Error("could not get channel's video IDs from db")
				return
			}

			// TODO: remove deleted videos from db too?
			allVideoIds, err := GetVideoIds(ctx, channelId)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				clog = log.With("error", err.Error())
				clog.Error("could not get channel's video IDs from yt api")
				return
			}

			videosToGet := inAnotInB(allVideoIds, storedVideoIds)
			// TODO: this can be more efficient with transactions, temp solution
			for _, vId := range videosToGet {
				vlog := clog.With("video_id", vId)
				video, err := GetVideo(vId)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					vlog = log.With("error", err.Error())
					vlog.Error("could not get video with yt_dlp")
					return
				}
				err = db.PutVideo(ctx, video)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					vlog = log.With("error", err.Error())
					vlog.Error("error putting video in db")
					return
				}
				vlog.Info("put video in db")
			}
			clog.Info("got all videos for channel")
		}

		// TODO: transcripts and text strings for videos that don't have them

		w.WriteHeader(http.StatusOK)
	})
}
