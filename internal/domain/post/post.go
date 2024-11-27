package post

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type PostID uuid.UUID

type Post struct {
	ID          PostID
	TeamID      uuid.UUID
	Title       string
	Content     string
	ScheduledAt *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func NewPost(teamID uuid.UUID, title, content string, scheduledAt *time.Time) (*Post, error) {
	if title == "" {
		return nil, errors.New("title cannot be empty")
	}
	if content == "" {
		return nil, errors.New("content cannot be empty")
	}

	return &Post{
		ID:          PostID(uuid.New()),
		TeamID:      teamID,
		Title:       title,
		Content:     content,
		ScheduledAt: scheduledAt,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}, nil
}

func (p *Post) Update(title, content string, scheduledAt *time.Time) error {
	if title == "" {
		return errors.New("title cannot be empty")
	}
	if content == "" {
		return errors.New("content cannot be empty")
	}

	p.Title = title
	p.Content = content
	p.ScheduledAt = scheduledAt
	p.UpdatedAt = time.Now()
	return nil
}
