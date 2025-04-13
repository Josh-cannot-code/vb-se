package youtube

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"go_server/database"
	"go_server/types"
	"net/http"
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

	videoId = "'" + videoId + "'"
	cmd := exec.Command(path, "python_scripts/get_video.py", videoId)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("stderr: %s, err: %w", string(stderr.Bytes()), err)
	}

	var vid types.YTDLPVideo
	err = json.Unmarshal(stdout.Bytes(), &vid)
	if err != nil {
		return nil, err
	}

	return &types.Video{
		ID:          vid.ID,
		VideoID:     vid.VideoID,
		Title:       vid.Title,
		Thumbnail:   vid.Thumbnail,
		Description: vid.Description,
		UploadDate:  vid.UploadDate.Time,
		Transcript:  vid.Transcript,
		URL:         vid.URL,
		ChannelID:   vid.ChannelID,
		ChannelName: vid.ChannelName,
	}, nil

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

func GetVideoTranscript(videoId string) (string, error) {
	path, err := exec.LookPath("python3")
	if err != nil {
		return "", err
	}

	// Get transcripts
	// TODO: quirk calling script
	cmd := exec.Command(path, "python_scripts/get_transcripts.py", "'"+videoId+"'")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		return "", fmt.Errorf("arg: %s, stderr: %s, err: %w", videoId, stderr.String(), err)
	}

	transcriptMap := make(map[string]string)
	err = json.Unmarshal(stdout.Bytes(), &transcriptMap)
	if err != nil {
		return "", err
	}
	return transcriptMap[videoId], nil
}

func GetVideoTranscripts(videoIds []string) (map[string]string, error) {
	if len(videoIds) == 0 {
		return make(map[string]string), nil
	}

	vidsArg := ""
	for _, vId := range videoIds {
		vidsArg += "," + "'" + vId + "'"
	}
	vidsArg = vidsArg[1:]

	path, err := exec.LookPath("python3")
	if err != nil {
		return nil, err
	}

	// Get transcripts
	cmd := exec.Command(path, "python_scripts/get_transcripts.py", vidsArg)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("arg: %s, stderr: %s, err: %w", vidsArg, string(stderr.Bytes()), err)
	}

	transcriptMap := make(map[string]string)
	err = json.Unmarshal(stdout.Bytes(), &transcriptMap)
	if err != nil {
		return nil, err
	}
	return transcriptMap, nil
}

// TODO: this should probably be async
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

			allVideoIds, err := GetVideoIds(ctx, channelId)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				clog = log.With("error", err.Error())
				clog.Error("could not get channel's video IDs from google youtube api")
				return
			}

			videosToGet := inAnotInB(allVideoIds, storedVideoIds)
			clog.Info("channel missing videos", "num_missing_vids", len(videosToGet))

			for _, vId := range videosToGet {
				vlog := clog.With("video_id", vId)
				video, err := GetVideo(vId)
				if err != nil {
					vlog.Error("could not get video with yt_dlp", "error", err.Error())
					continue
				}
				if video == nil {
					vlog.Warn("video likely does not exist")
					video = &types.Video{
						ID: vId,
					}
				}

				transcript, err := GetVideoTranscript(vId)
				if err != nil {
					vlog.Error("could not get video transcript with transcript api", "error", err.Error())
					continue
				}
				video.Transcript = transcript

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

		/*
			noTransVIds, err := db.GetNoTranscriptVideoIds(ctx)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				log.Error("could not get videos without transcripts from db", "error", err.Error())
				return
			}

			log.Info("getting missing transcripts", "num_missing_transcripts", len(noTransVIds))

			// Paginate request to get transcripts
			i := 0
			pageSize := 10
			for i < len(noTransVIds) {
				var k int
				if i+pageSize >= len(noTransVIds) {
					k = len(noTransVIds)
				} else {
					k = i + pageSize
				}

				glog := log.WithGroup("transcript page")
				glog.Info("getting transcripts", "start", i, "end", k, "total", len(noTransVIds))

				transcriptMap, err := GetVideoTranscripts(noTransVIds[i:k])
				i = k // Increment

				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					glog.Error("could not get video transcripts", "error", err.Error())
					return
				}

				err = db.UpdateVideoTranscripts(ctx, transcriptMap)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					glog.Error("could not update video transcripts", "error", err.Error())
					return
				}
				glog.Info("updated transcripts in db")
			}

			log.Info("all missing transcripts updated")
			log.Info("updating missing video text data")
			err = db.UpdateVideoTextData(ctx)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				log.Error("could not update video text data", "error", err.Error())
				return

			}
			log.Info("missing video text data updated")
		*/
		w.WriteHeader(http.StatusOK)
	})
}
