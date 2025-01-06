package post

import (
	"context"
	"fmt"
	"time"

	"github.com/pedrodcsjostrom/opencm/internal/interfaces/api/http/middlewares"
	"golang.org/x/sync/errgroup"
)

//go:generate mockery --name=Service --case=underscore --inpackage
type Service interface {
	CreatePost(
		ctx context.Context,
		projectID, title, textContent string,
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
	AddToProjectQueue(ctx context.Context, projectID, postID string) error
	GetProjectQueuedPosts(ctx context.Context, projectID string) ([]*Post, error)
	MovePostInQueue(ctx context.Context, projectID string, currentIndex, newIndex int) error
	DequeuePostsToPublish(ctx context.Context, projectID string) ([]*QPost, error)
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
	isIdea bool,
	scheduledAt time.Time,
) (*Post, error) {
	userID := ctx.Value(middlewares.UserIDKey).(string)

	p, err := NewPost(
		projectID,
		userID,
		title,
		textContent,
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

func (s *service) AddToProjectQueue(ctx context.Context, projectID, postID string) error {
	var (
		p     *Post
		queue *Queue
	)

	g, gCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		var err error
		p, err = s.repo.FindByID(gCtx, postID)
		return err
	})

	g.Go(func() error {
		var err error
		queue, err = s.repo.GetProjectPostQueue(gCtx, projectID)
		return err
	})

	if err := g.Wait(); err != nil {
		return err
	}

	fmt.Println("queue", queue)
	if p == nil {
		return ErrPostNotFound
	}
	if p.Status == string(PostStatusPublished) {
		return ErrPostAlreadyPublished
	}
	if p.IsIdea {
		return ErrPostIsIdea
	}
	if queue.Contains(p.ID) {
		return ErrPostAlreadyInQueue
	}

	p.Status = string(PostStatusQueued)
	p.ScheduledAt = time.Time{}

	err := s.repo.Update(ctx, p)
	if err != nil {
		return err
	}

	return s.repo.AddToProjectQueue(ctx, projectID, postID)
}

func (s *service) GetProjectQueuedPosts(ctx context.Context, projectID string) ([]*Post, error) {
	q, err := s.repo.GetProjectPostQueue(ctx, projectID)
	if err != nil {
		return nil, err
	}
	if q.IsEmpty() {
		return []*Post{}, nil
	}
	qp, err := s.repo.GetProjectQueuedPosts(ctx, projectID, q.Arr())
	if err != nil {
		return nil, err
	}
	qp = sortPostsByQueue(qp, q)
	return qp, nil
}

func sortPostsByQueue(posts []*Post, queue *Queue) []*Post {
	sortedPosts := make([]*Post, 0)
	// Double loop but it's fine since the queue is small
	for _, postID := range queue.Arr() {
		for _, post := range posts {
			if post.ID == postID {
				sortedPosts = append(sortedPosts, post)
			}
		}
	}
	return sortedPosts
}

func (s *service) MovePostInQueue(ctx context.Context, projectID string, currentIndex, newIndex int) error {
	q, err := s.repo.GetProjectPostQueue(ctx, projectID)
	if err != nil {
		return err
	}
	if q.IsEmpty() {
		return nil
	}
	q.Move(currentIndex, newIndex)
	return s.repo.UpdateProjectPostQueue(ctx, projectID, q.Arr())
}

func (s *service) DequeuePostsToPublish(ctx context.Context, projectID string) ([]*QPost, error) {
	q, err := s.repo.GetProjectPostQueue(ctx, projectID)
	if err != nil {
		return nil, err
	}
	if q.IsEmpty() {
		return nil, nil
	}
	postID := q.Shift()
	err = s.repo.UpdateProjectPostQueue(ctx, projectID, q.Arr())
	if err != nil {
		return nil, err
	}
	return s.repo.GetPostsForPlatformPublishQueue(ctx, postID)
}