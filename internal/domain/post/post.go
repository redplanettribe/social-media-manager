package post

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Represents the status of a post
type PostStatus string

// Possible statuses of a parent high level post
const (
	PostStatusDraft     PostStatus = "draft"
	PostStatusQueued    PostStatus = "queued"
	PostStatusScheduled PostStatus = "scheduled"
	PostStatusPublished PostStatus = "published"
	PostStatusFailed    PostStatus = "failed"
	PostStatusArchived  PostStatus = "archived"
)

type PublisherPostStatus string

// Status of a post in the publisher
const (
	PublisherPostStatusReady      PublisherPostStatus = "ready"
	PublisherPostStatusProcessing PublisherPostStatus = "processing"
	PublisherPostStatusPublished  PublisherPostStatus = "published"
	PublisherPostStatusFailed     PublisherPostStatus = "failed"
)

type PostType string

// Possible types of a post
const (
	PostTypeUndefined  PostType = "undefined"
	PostTypeText       PostType = "text"
	PostTypeMedia      PostType = "media"
	PostTypePoll       PostType = "poll"
	PostTypeShortVideo PostType = "short_video"
	// ... add other types as necessary
)

// Error messages
var (
	ErrProjectNotFound       = errors.New("project not found")
	ErrPostNotFound          = errors.New("post not found")
	ErrPublisherNotInProject = errors.New("publisher not in project")
	ErrPostScheduledTime     = errors.New("post scheduled time is in the past")
	ErrPostAlreadyInQueue    = errors.New("post already in queue")
	ErrPostAlreadyPublished  = errors.New("post already published")
	ErrPostIsIdea            = errors.New("post is an idea")
)

type Post struct {
	ID          string    `json:"id"`
	ProjectID   string    `json:"project_id"`
	Title       string    `json:"title"`
	Type        PostType  `json:"type"`
	TextContent string    `json:"text_content"`
	IsIdea      bool      `json:"is_idea"`
	Status      string    `json:"status"`
	CreatedBy   string    `json:"created_by"`
	ScheduledAt time.Time `json:"scheduled_at"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type PublishPost struct {
	// Post fields
	*Post
	// Additional fields
	Secrets       string `json:"secrets"`
	Platform      string `json:"platform"`
	PublishStatus string `json:"publish_status"`
}

func NewPost(
	projectID, userID string,
	title, postType, content string,
	isIdea bool,
	scheduledAt time.Time,
) (*Post, error) {
	if title == "" {
		return nil, errors.New("title cannot be empty")
	}
	if content == "" {
		return nil, errors.New("content cannot be empty")
	}
	if projectID == "" {
		return nil, errors.New("projectID cannot be empty")
	}
	if userID == "" {
		return nil, errors.New("userID cannot be empty")
	}

	return &Post{
		ID:          uuid.New().String(),
		ProjectID:   projectID,
		Title:       title,
		Type:        PostType(postType),
		TextContent: content,
		IsIdea:      isIdea,
		Status:      string(PostStatusDraft),
		CreatedBy:   userID,
		ScheduledAt: scheduledAt,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}, nil
}

func (p *Post) Update(title, content string, scheduledAt time.Time) error {
	if title == "" {
		return errors.New("title cannot be empty")
	}
	if content == "" {
		return errors.New("content cannot be empty")
	}

	p.Title = title
	p.TextContent = content
	p.ScheduledAt = scheduledAt
	p.UpdatedAt = time.Now()
	return nil
}
