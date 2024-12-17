package types

import (
	"encoding/json"
	"errors"
	"time"
)

type YTDLPTime struct {
	time.Time
}

func (u *YTDLPTime) UnmarshalJSON(b []byte) error {
	var timestamp string
	err := json.Unmarshal(b, &timestamp)
	if err != nil {
		return err
	}
	u.Time, err = time.Parse("20060102", timestamp)
	if err != nil {
		return err
	}
	return nil
}

func (u *YTDLPTime) Scan(value interface{}) error {
	time, ok := value.(time.Time)
	if !ok {
		return errors.New("incompatable type")
	}
	u.Time = time
	return nil
}

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

type YTDLPVideo struct {
	ID          string    `json:"id"`
	VideoID     string    `json:"video_id"`
	Title       string    `json:"title"`
	Thumbnail   string    `json:"thumbnail"`
	ChannelID   string    `json:"channel_id"`
	Description string    `json:"description"`
	UploadDate  YTDLPTime `json:"upload_date"`
	URL         string    `json:"url"`
	ChannelName string    `json:"channel_name"`
	Transcript  string    `json:"transcript"`
}

type VCardData struct {
	TitleSnippet string
	DescSnippet  string
	TransSnippet string
	Video        Video
}
