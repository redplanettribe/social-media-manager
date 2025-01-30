package post

import (
	"context"
	"time"

	"github.com/pedrodcsjostrom/opencm/internal/interfaces/api/http/middlewares"
	"golang.org/x/sync/errgroup"
)

//go:generate mockery --name=Service --case=underscore --inpackage
type Service interface {
	CreatePost(
		ctx context.Context,
		projectID, title, postType, textContent string,
		isIdea bool,
		scheduledAt time.Time) (*Post, error)
	GetPost(ctx context.Context, id string) (*PostResponse, error)
	ListProjectPosts(ctx context.Context, projectID string) ([]*Post, error)
	ArchivePost(ctx context.Context, id string) error
	DeletePost(ctx context.Context, id string) error
	AddSocialMediaPublisher(ctx context.Context, projectID, postID, publisherID string) error
	GetSocialMediaPublishers(ctx context.Context, postID string) ([]string, error)
	FindScheduledReadyPosts(ctx context.Context, offset, chunkSize int) ([]*PublishPost, error)
	GetPostToPublish(ctx context.Context, id string) (*PublishPost, error)
	SchedulePost(ctx context.Context, id string, scheduledAt time.Time) error
	UnschedulePost(ctx context.Context, id string) error
	AddToProjectQueue(ctx context.Context, projectID, postID string) error
	RemoveFromProjectQueue(ctx context.Context, projectID, postID string) error
	GetProjectQueuedPosts(ctx context.Context, projectID string) ([]*Post, error)
	MovePostInQueue(ctx context.Context, projectID string, currentIndex, newIndex int) error
	DequeuePostsToPublish(ctx context.Context, projectID string) ([]*PublishPost, error)
	GetAvailablePostTypes() []string
	UpdatePostStatus(ctx context.Context, id string, status PostStatus) error
	UpdatePublishPostStatus(ctx context.Context, postID, platformID string, status PublishPostStatus) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) CreatePost(
	ctx context.Context,
	projectID, title, postType, textContent string,
	isIdea bool,
	scheduledAt time.Time,
) (*Post, error) {
	userID := ctx.Value(middlewares.UserIDKey).(string)

	if !PostType(postType).IsValid() {
		return nil, ErrInvalidPostType
	}

	p, err := NewPost(
		projectID,
		userID,
		title,
		postType,
		textContent,
		isIdea,
		scheduledAt)
	if err != nil {
		return &Post{}, err
	}

	if p.IsIdea {
		err := s.repo.AddToProjectIdeaQueue(ctx, projectID, p.ID)
		if err != nil {
			return &Post{}, err
		}
	}

	err = s.repo.Save(ctx, p)
	if err != nil {
		return &Post{}, err
	}
	return p, nil
}

func (s *service) GetPost(ctx context.Context, id string) (*PostResponse, error) {
	var (
		p               *Post
		linkedPlatforms []Platform
		g               errgroup.Group
	)

	g.Go(func() error {
		var err error
		p, err = s.repo.FindByID(ctx, id)
		if p == nil {
			return ErrPostNotFound
		}
		return err
	})

	g.Go(func() error {
		var err error
		linkedPlatforms, err = s.repo.GetSocialMediaPlatforms(ctx, id)
		return err
	})

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return &PostResponse{
		Post:            p,
		LinkedPlatforms: linkedPlatforms,
	}, nil
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

func (s *service) GetSocialMediaPublishers(ctx context.Context, postID string) ([]string, error) {
	return s.repo.GetSocialMediaPublishersIDs(ctx, postID)
}

func (s *service) FindScheduledReadyPosts(ctx context.Context, offset, chunkSize int) ([]*PublishPost, error) {
	return s.repo.FindScheduledReadyPosts(ctx, offset, chunkSize)
}

func (s *service) GetPostToPublish(ctx context.Context, postID string) (*PublishPost, error) {
	return s.repo.GetPostToPublish(ctx, postID)
}

func (s *service) SchedulePost(ctx context.Context, id string, scheduletAt time.Time) error {
	var (
		p         *Post
		platforms []Platform
	)

	g, gCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		var err error
		p, err = s.repo.FindByID(gCtx, id)
		if err != nil {
			return err
		}
		if p == nil {
			return ErrPostNotFound
		}

		if p.Status != string(PostStatusDraft) {
			return ErrPostNotDraft
		}
		return nil
	})

	g.Go(func() error {
		var err error
		platforms, err = s.repo.GetSocialMediaPlatforms(gCtx, id)
		if len(platforms) == 0 {
			return ErrPostNotLinkedToAnyPlatform
		}
		return err
	})

	if err := g.Wait(); err != nil {
		return err
	}

	if scheduletAt.Before(time.Now().UTC()) {
		return ErrPostScheduledTime
	}
	return s.repo.SchedulePost(ctx, id, scheduletAt)
}

func (s *service) UnschedulePost(ctx context.Context, id string) error {
	p, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if p == nil {
		return ErrPostNotFound
	}
	if p.Status != string(PostStatusScheduled) {
		return ErrPostNotScheduled
	}
	return s.repo.UnschedulePost(ctx, id)
}

func (s *service) AddToProjectQueue(ctx context.Context, projectID, postID string) error {
	var (
		p         *Post
		queue     *Queue
		platforms []Platform
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

	g.Go(func() error {
		var err error
		platforms, err = s.repo.GetSocialMediaPlatforms(gCtx, postID)
		if len(platforms) == 0 {
			return ErrPostNotLinkedToAnyPlatform
		}
		return err
	})

	if err := g.Wait(); err != nil {
		return err
	}

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

func (s *service) RemoveFromProjectQueue(ctx context.Context, projectID, postID string) error {
	var (
		p *Post
	)

	g, gCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		var err error
		p, err = s.repo.FindByID(gCtx, postID)
		return err
	})

	g.Go(func() error {
		var err error
		queue, err := s.repo.GetProjectPostQueue(gCtx, projectID)
		if err != nil {
			return err
		}
		if !queue.Contains(postID) {
			return ErrPostNotInQueue
		}
		return err
	})

	if err := g.Wait(); err != nil {
		return err
	}

	if p == nil {
		return ErrPostNotFound
	}
	if p.Status == string(PostStatusPublished) {
		return ErrPostAlreadyPublished
	}
	if p.IsIdea {
		return ErrPostIsIdea
	}

	p.Status = string(PostStatusDraft)

	g2, gCtx2 := errgroup.WithContext(ctx)

	g2.Go(func() error {
		return s.repo.Update(gCtx2, p)
	})

	g2.Go(func() error {
		return s.repo.RemoveFromProjectQueue(gCtx2, projectID, postID)
	})

	return g2.Wait()
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

func (s *service) DequeuePostsToPublish(ctx context.Context, projectID string) ([]*PublishPost, error) {
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
	return s.repo.GetPostsForPublishQueue(ctx, postID)
}

func (s *service) GetAvailablePostTypes() []string {
	return []string{
		PostTypeCarousel.String(),
		PostTypeImage.String(),
		PostTypeMultiImage.String(),
		PostTypeShortVideo.String(),
		PostTypeText.String(),
		PostTypeVideo.String(),
		PostTypeDocument.String(),
	}
}

func (s *service) UpdatePostStatus(ctx context.Context, id string, status PostStatus) error {
	p, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if p == nil {
		return ErrPostNotFound
	}
	p.Status = string(status)
	return s.repo.Update(ctx, p)
}

func (s *service) UpdatePublishPostStatus(ctx context.Context, postID, platformID string, status PublishPostStatus) error {
	return s.repo.UpdatePublishPostStatus(ctx, postID, platformID, string(status))
}
