package media

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

const (
	thumbnailPrefix = "thumbnail_"
)

var (
	ErrInvalidMedia                 = errors.New("invalid media")
	ErrPostDoesNotBelongToProject   = errors.New("post does not belong to project")
	ErrMediaDoesNotBelongToPost     = errors.New("media does not belong to post")
	ErrPlatformNotEnabledForProject = errors.New("platform not enabled for project")
	ErrPostNotLinkedToPlatform      = errors.New("post not linked to platform")
	ErrMediaAlreadyLinkedToPost     = errors.New("media already linked to post")
	ErrFileAlreadyExists            = errors.New("file already exists")
	ErrFailedToAnalyzeMedia         = errors.New("failed to analyze media")
)

type Media struct {
	Data      []byte
	Thumbnail *Media
	*MetaData
}

type MediaType string

const (
	MediaTypeImage      MediaType = "image"
	MediaTypeVideo      MediaType = "video"
	MediaTypeShortVideo MediaType = "short_video"
	MediaTypeDocument   MediaType = "document"
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
	Size      int       `json:"size"` // in bytes
	AltText   string    `json:"alt_text"`
	AddedBy   string    `json:"added_by"`
	CreatedAt time.Time `json:"created_at"`
}

func (m *MetaData) IsImage() bool {
	return m.Type == MediaTypeImage
}

func (m *MetaData) IsVideo() bool {
	return m.Type == MediaTypeVideo
}

func getThumbnailName(name string) string {
	return thumbnailPrefix + name + "." + ThumbnailFormat
}

func NewMetadata(postID, userID, fileName, altText string, data []byte, mediaInfo *MediaInfo) (*MetaData, error) {
	return &MetaData{
		ID:        uuid.New().String(),
		PostID:    postID,
		Filename:  fileName,
		Type:      mediaInfo.Type,
		Width:     mediaInfo.Width,
		Height:    mediaInfo.Height,
		Length:    mediaInfo.Length,
		Format:    mediaInfo.Format,
		Size:      mediaInfo.Size,
		AltText:   altText,
		AddedBy:   userID,
		CreatedAt: time.Now(),
	}, nil
}

type DownloadMediaData struct {
	Url          *string `json:"url"`
	UrlThumbnail *string `json:"url_thumbnail"`
	*MetaData
}
