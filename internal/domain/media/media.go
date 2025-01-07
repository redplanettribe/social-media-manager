package media

import (
	"time"

	"github.com/google/uuid"
)

type MediaType string

const (
	MediaTypeImage      MediaType = "image"
	MediaTypeVideo      MediaType = "video"
	MediaTypeShortVideo MediaType = "short_video"
)

type MetaData struct {
	ID           string    `json:"id"`
	PostID       string    `json:"post_id"`
	Type         MediaType `json:"media_type"`
	Url          string    `json:"media_url"`
	ThumbnailUrl string    `json:"thumbnail_url"`
	Width        int       `json:"width"`
	Height       int       `json:"height"`
	Length       int       `json:"length"`
	AddedBy      string    `json:"added_by"`
	CreatedAt    time.Time `json:"created_at"`
}

func NewMetadata(postID, userID, fileName string, data []byte) (*MetaData, error) {
	analyzer, err := GetAnalyzer(fileName)
	if err != nil {
		return nil, err
	}
	mediaInfo, err := analyzer.Analyze(data)
	if err != nil {
		return nil, err
	}

	// Not adding the URL and Thumbnail URL here, as it will be added by the object repository
	return &MetaData{
		ID:           uuid.New().String(),
		PostID:       postID,
		Type:         mediaInfo.Type,
		Url:          "",
		ThumbnailUrl: "",
		Width:        mediaInfo.Width,
		Height:       mediaInfo.Height,
		Length:       mediaInfo.Length,
		AddedBy:      userID,
		CreatedAt:    time.Now(),
	}, nil
}

type Urls struct {
	Url          string
	ThumbnailUrl string
}