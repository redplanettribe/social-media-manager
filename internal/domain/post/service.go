package post

import (
	"context"
	"time"

	"github.com/pedrodcsjostrom/opencm/internal/interfaces/api/http/middlewares"
)

//go:generate mockery --name=Service --case=underscore --inpackage
type Service interface {
	CreatePost(
		ctx context.Context,
		projectID, title, textContent string,
		imageURLs, videoURLs []string,
		isIdea bool,
		scheduledAt time.Time) (*Post, error)
	GetPost(ctx context.Context, id string) (*Post, error)
	ListProjectPosts(ctx context.Context, projectID string) ([]*Post, error)
	ArchivePost(ctx context.Context, id string) error
	DeletePost(ctx context.Context, id string) error
	AddSocialMediaPublisher(ctx context.Context, projectID, postID, publisherID string) error
	FindScheduledReadyPosts(ctx context.Context, offset, chunkSize int) ([]*QPost, error)
	GetQueuePost(ctx context.Context, id string) (*QPost, error)
	SchedulePost(ctx context.Context, id string, scheduled_at time.Time) error
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

	err = s.repo.Save(ctx, p)
	if err != nil {
		return &Post{}, err
	}
	return p, nil
}

func (s *service) GetPost(ctx context.Context, id string) (*Post, error) {
	p, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return &Post{}, err
	}
	if p == nil {
		return &Post{}, ErrPostNotFound
	}
	return p, nil
}

func (s *service) ListProjectPosts(ctx context.Context, projectID string) ([]*Post, error) {
	return s.repo.FindByProjectID(ctx, projectID)
}

func (s *service) ArchivePost(ctx context.Context, id string) error {
	p, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if p == nil {
		return ErrPostNotFound
	}
	return s.repo.ArchivePost(ctx, id)
}

func (s *service) DeletePost(ctx context.Context, id string) error {
	p, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if p == nil {
		return ErrPostNotFound
	}
	return s.repo.DeletePost(ctx, id)
}

func (s *service) AddSocialMediaPublisher(ctx context.Context, projectID, postID, publisherID string) error {
	ok, err := s.repo.IsPublisherPlatformEnabledForProject(ctx, projectID, publisherID)
	if err != nil {
		return err
	} else if !ok {
		return ErrPublisherNotInProject
	}	
	p, err := s.repo.FindByID(ctx, postID)
	if err != nil {
		return err
	}
	if p == nil {
		return ErrPostNotFound
	}
	return s.repo.AddSocialMediaPublisher(ctx, postID, publisherID)
}

func (s *service) FindScheduledReadyPosts(ctx context.Context, offset, chunkSize int) ([]*QPost, error) {
	return s.repo.FindScheduledReadyPosts(ctx, offset, chunkSize)
}

func (s *service) GetQueuePost(ctx context.Context, id string) (*QPost, error) {
	// TODO: Implement this
	return &QPost{}, nil
}

func (s *service) SchedulePost(ctx context.Context, id string, sheduled_at time.Time) error {
	if sheduled_at.Before(time.Now()) {
		return ErrPostScheduledTime	
	}
	return s.repo.SchedulePost(ctx, id, sheduled_at)
}