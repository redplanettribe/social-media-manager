package post

import (
	"context"
	"time"

	"github.com/pedrodcsjostrom/opencm/internal/interfaces/api/http/middlewares"
)

type Service interface {
	CreatePost(
		ctx context.Context,
		projectID, title, textContent string,
		imageURLs, videoURLs []string,
		isIdea bool,
		scheduledAt time.Time) (*Post, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) CreatePost(
	ctx context.Context,
	projectID, title, textContent string,
	imageURLs, videoURLs []string,
	isIdea bool,
	scheduledAt time.Time,
) (*Post, error) {
	userID := ctx.Value(middlewares.UserIDKey).(string)

	p, err := NewPost(
		projectID,
		userID,
		title,
		textContent,
		imageURLs,
		videoURLs,
		isIdea,
		scheduledAt)
	if err != nil {
		return &Post{}, err
	}

	err= s.repo.Save(ctx, p)
	if err != nil {
		return &Post{}, err
	}
	return p, nil
}
