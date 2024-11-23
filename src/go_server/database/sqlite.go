package database

import (
	"database/sql"
	"go_server/types"
	"log"

	"golang.org/x/net/context"
)

type SqLiteConnection struct {
	db *sql.DB
}

func NewSqLiteConnection(db *sql.DB) *SqLiteConnection {
	return &SqLiteConnection{
		db: db,
	}
}

func (c SqLiteConnection) PutVideo(ctx context.Context, vid *types.Video) error {
	sqlStatement := `INSERT INTO videos (
            title,
            upload_date,
            url,
            thumbnail,
            transcript,
            description,
            video_id,
            channel_id,
            channel_name
        ) VALUES (
            $1,
            $2,
            $3,
            $4,
            $5,
            $6,
            $7,
            $8,
            $9
        )`
	result, err := c.db.Exec(
		sqlStatement,
		vid.Title,
		vid.UploadDate.Time.Format("2006-01-02 15:04:05"),
		vid.URL,
		vid.Thumbnail,
		vid.Transcript,
		vid.Description,
		vid.ID,
		vid.ChannelID,
		vid.ChannelName,
	)
	if err != nil {
		return err
	}
	log.Printf("result: %v\n", result)
	return nil
}

func (c SqLiteConnection) GetChannelIds(ctx context.Context) ([]string, error) {
	sqlStatement := `SELECT channel_id FROM channels`
	rows, err := c.db.Query(sqlStatement)
	if err != nil {
		return nil, err
	}

	var channelIds []string
	for rows.Next() {
		var curChannelId string
		err = rows.Scan(&curChannelId)
		if err != nil {
			// TODO: log.Error here
			return nil, err
		}
		channelIds = append(channelIds, curChannelId)
	}
	return channelIds, nil
}

func (c SqLiteConnection) GetVideoIds(ctx context.Context, channelId string) ([]string, error) {
	sqlStatement := `SELECT video_id FROM videos WHERE channel_id = $1`
	rows, err := c.db.Query(sqlStatement, channelId)
	if err != nil {
		return nil, err
	}
	var videoIds []string
	for rows.Next() {
		var curVideoId string
		err = rows.Scan(&curVideoId)
		if err != nil {
			// TODO: log.Error here? maybe idk bout these
			return nil, err
		}
		videoIds = append(videoIds, curVideoId)
	}
	return videoIds, nil
}
