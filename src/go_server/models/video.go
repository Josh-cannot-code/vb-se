package models

import (
	"time"
)

type Video struct {
	ID          string    `json:"id"`
	VideoID     string    `json:"video_id"`
	Title       string    `json:"title"`
	Thumbnail   string    `json:"thumbnail"`
	ChannelID   string    `json:"channel_id"`
	Description string    `json:"description"`
	UploadDate  time.Time `json:"upload_date"`
	URL         string    `json:"url"`
	ChannelName string    `json:"channel_name"`
	Transcript  string    `json:"transcript"`
}
