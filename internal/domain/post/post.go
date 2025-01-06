package post

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Represents the status of a post
type PostStatus string

const (
	PostStatusDraft     PostStatus = "draft"
	PostStatusQueued    PostStatus = "queued"
	PostStatusScheduled PostStatus = "scheduled"
	PostStatusPublished PostStatus = "published"
	PostStatusFailed    PostStatus = "failed"
	PostStatusArchived  PostStatus = "archived"
)

type PublisherPostStatus string

const (
	PublisherPostStatusReady      PublisherPostStatus = "ready"
	PublisherPostStatusProcessing PublisherPostStatus = "processing"
	PublisherPostStatusPublished  PublisherPostStatus = "published"
	PublisherPostStatusFailed     PublisherPostStatus = "failed"
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
	ID          string        `json:"id"`
	ProjectID   string        `json:"project_id"`
	Title       string        `json:"title"`
	TextContent string        `json:"text_content"`
	IsIdea      bool          `json:"is_idea"`
	Status      string        `json:"status"`
	CreatedBy   string        `json:"created_by"`
	ScheduledAt time.Time     `json:"scheduled_at"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}

type QPost struct {
	// Post fields
	ID          string        `json:"id"`
	ProjectID   string        `json:"project_id"`
	Title       string        `json:"title"`
	TextContent string        `json:"text_content"`
	IsIdea      bool          `json:"is_idea"`
	Status      string        `json:"status"`
	CreatedBy   string        `json:"created_by"`
	ScheduledAt time.Time     `json:"scheduled_at"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
	// Additional fields
	ApiKey        string
	Platform      string `json:"platform"`
	PublishStatus string `json:"publish_status"`
}

func NewPost(
	projectID, userID string,
	title, content string,
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
