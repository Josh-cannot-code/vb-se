package types

import (
	"encoding/json"
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

func (u *YTDLPTime) MarshalJSON() ([]byte, error) {
	value, err := json.Marshal(u.Time)
	if err != nil {
		return nil, err
	}
	return value, nil
}

type Video struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Thumbnail   string    `json:"thumbnail"`
	ChannelID   string    `json:"channel_id"`
	Description string    `json:"description"`
	UploadDate  YTDLPTime `json:"upload_date"`
	URL         string    `json:"url"`
	ChannelName string    `json:"channel_name"`
	Transcript  string
}
