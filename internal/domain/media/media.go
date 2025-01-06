package media

import "time"

type MediaType string

const (
	MediaTypeImage      MediaType = "image"
	MediaTypeVideo      MediaType = "video"
	MediaTypeShortVideo MediaType = "short_video"
)

type Media struct {
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
