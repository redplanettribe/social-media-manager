package media

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidMedia= errors.New("invalid media")
	ErrPostDoesNotBelongToProject= errors.New("post does not belong to project")
	ErrMediaDoesNotBelongToPost= errors.New("media does not belong to post")
	ErrPlatformNotEnabledForProject= errors.New("platform not enabled for project")
	ErrPostNotLinkedToPlatform= errors.New("post not linked to platform")
)
type Media struct {
	Data []byte
	*MetaData
}

type MediaType string

const (
	MediaTypeImage      MediaType = "image"
	MediaTypeVideo      MediaType = "video"
	MediaTypeShortVideo MediaType = "short_video"
)

type MetaData struct {
	ID        string    `json:"id"`
	PostID    string    `json:"post_id"`
	Filename  string    `json:"filename"`
	Type      MediaType `json:"media_type"`
	Format    string    `json:"format"`
	Width     int       `json:"width"`
	Height    int       `json:"height"`
	Length    int       `json:"length"`
	AddedBy   string    `json:"added_by"`
	CreatedAt time.Time `json:"created_at"`
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

	return &MetaData{
		ID:        uuid.New().String(),
		PostID:    postID,
		Filename:  fileName,
		Type:      mediaInfo.Type,
		Width:     mediaInfo.Width,
		Height:    mediaInfo.Height,
		Length:    mediaInfo.Length,
		Format:    mediaInfo.Format,
		AddedBy:   userID,
		CreatedAt: time.Now(),
	}, nil
}
