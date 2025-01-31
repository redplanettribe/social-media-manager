package publisher

import (
	"context"
	"fmt"

	"github.com/pedrodcsjostrom/opencm/internal/domain/media"
	post "github.com/pedrodcsjostrom/opencm/internal/domain/post"
	"github.com/pedrodcsjostrom/opencm/internal/infrastructure/encrypting"
	"golang.org/x/sync/errgroup"
)

//go:generate mockery --name=Service --case=underscore --inpackage
type Service interface {
	GetAvailableSocialNetworks(ctx context.Context) ([]Platform, error)
	PublishPostToAssignedSocialNetworks(ctx context.Context, projecID, postID string) error
	ValidatePostForAssignedSocialNetworks(ctx context.Context, projecID, postID string) error
	PublishPostToSocialNetwork(ctx context.Context, projectID, postID, platformID string) error
	ValidatePostForSocialNetwork(ctx context.Context, projectID, postID, platformID string) error
	Authenticate(ctx context.Context, platformID, projectID, userID, code string) error
	GetPublishPostInfo(ctx context.Context, projectID, postID, platformID string) (*PublishPostInfo, error)
}

type service struct {
	repo             Repository
	publisherFactory PublisherFactory
	encrypter        encrypting.Encrypter
	postService      post.Service
	mediaService     media.Service
}

func NewService(r Repository, e encrypting.Encrypter, pf PublisherFactory, ps post.Service, m media.Service) Service {
	return &service{
		repo:             r,
		publisherFactory: pf,
		encrypter:        e,
		postService:      ps,
		mediaService:     m,
	}
}

func (s *service) GetPublishPostInfo(ctx context.Context, projectID, postID, platformID string) (*PublishPostInfo, error) {
	var (
		publishPost *post.PublishPost
		media       []*media.Media
		g           errgroup.Group
	)

	g.Go(func() error {
		var err error
		publishPost, err = s.postService.GetPostToPublish(ctx, postID)
		return err
	})

	g.Go(func() error {
		var err error
		media, err = s.mediaService.GetMediaForPost(ctx, projectID, postID, platformID)
		return err
	})

	if err := g.Wait(); err != nil {
		return nil, err
	}
	if publishPost == nil {
		return nil, post.ErrPostNotFound
	}
	if publishPost.ProjectID != projectID {
		return nil, post.ErrPostNotInProject
	}

	return &PublishPostInfo{
		Post:  publishPost,
		Media: media,
	}, nil
}

func (s *service) Authenticate(ctx context.Context, platformID, projectID, userID, code string) error {

	var (
		isEnabled   bool
		userSecrets string
	)

	g, gCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		var err error
		isEnabled, err = s.repo.IsSocialNetworkEnabledForProject(gCtx, projectID, platformID)
		return err
	})

	g.Go(func() error {
		var err error
		userSecrets, err = s.repo.GetUserPlatformSecrets(gCtx, platformID, userID)
		return err
	})

	if err := g.Wait(); err != nil {
		return err
	}

	if !isEnabled {
		return ErrSocialPlatformNotEnabledForProject
	}

	publisher, err := s.publisherFactory.Create(platformID, userSecrets)
	if err != nil {
		return err
	}

	encryptedSecrets, toknTtl, err := publisher.Authenticate(ctx, code)
	if err != nil {
		return err
	}

	return s.repo.SetUserPlatformAuthSecretsWithTTL(ctx, platformID, userID, encryptedSecrets, toknTtl)
}

func (s *service) GetAvailableSocialNetworks(ctx context.Context) ([]Platform, error) {
	return s.repo.FindAll(ctx)
}

func (s *service) PublishPostToAssignedSocialNetworks(ctx context.Context, projectID, postID string) error {
	publishers, err := s.postService.GetSocialMediaPublishers(ctx, postID)
	if err != nil {
		return err
	}

	if len(publishers) == 0 {
		return ErrNoPublishersAssigned
	}

	// Create channel to collect results
	type publishResult struct {
		platformID string
		err        error
	}
	results := make(chan publishResult, len(publishers))

	// Launch goroutines for each publisher
	g, gCtx := errgroup.WithContext(ctx)
	for _, publisherID := range publishers {
		pid := publisherID
		g.Go(func() error {
			err := s.PublishPostToSocialNetwork(gCtx, projectID, postID, pid)
			results <- publishResult{pid, err}
			return nil // Don't propagate errors through errgroup
		})
	}

	g.Wait()
	close(results)

	var failures int
	failedPlatforms := make([]string, 0)
	for result := range results {
		if result.err != nil {
			failures++
			failedPlatforms = append(failedPlatforms, result.platformID)
		}
	}

	// Update post status based on results
	if failures == len(publishers) {
		// All publishers failed
		if err := s.postService.UpdatePostStatus(ctx, postID, post.PostStatusFailed); err != nil {
			return fmt.Errorf("failed to update post status: %w", err)
		}
		return fmt.Errorf("all publishers failed: %v", failedPlatforms)
	} else if failures > 0 {
		// Some publishers failed
		if err := s.postService.UpdatePostStatus(ctx, postID, post.PostStatusPartialyPublished); err != nil {
			return fmt.Errorf("failed to update post status: %w", err)
		}
		return fmt.Errorf("some publishers failed: %v", failedPlatforms)
	}

	// All successful
	return s.postService.UpdatePostStatus(ctx, postID, post.PostStatusPublished)
}

func (s *service) ValidatePostForAssignedSocialNetworks(ctx context.Context, projectID, postID string) error {
	publishers, err := s.postService.GetSocialMediaPublishers(ctx, postID)
	if err != nil {
		return err
	}

	if len(publishers) == 0 {
		return ErrNoPublishersAssigned
	}

	g, gCtx := errgroup.WithContext(ctx)
	for _, publisherID := range publishers {
		pid := publisherID
		g.Go(func() error {
			err := s.ValidatePostForSocialNetwork(gCtx, projectID, postID, pid)
			return fmt.Errorf("failed to validate post for %s: %w", pid, err)
		})
	}

	if err := g.Wait(); err != nil {
		return err
	}

	return nil
}

func (s *service) PublishPostToSocialNetwork(ctx context.Context, projectID, postID, platformID string) error {
	var (
		isEnabled     bool
		publishPost   *post.PublishPost
		media         []*media.Media
		secrets       string
		defaultUserID string
		g             errgroup.Group
	)

	defaultUserID, err := s.repo.GetDefaultUserID(ctx, projectID)
	if err != nil {
		return err
	}
	if defaultUserID == "" {
		return ErrDefaultUserNotSet
	}

	g.Go(func() error {
		var err error
		isEnabled, err = s.repo.IsSocialNetworkEnabledForProject(ctx, projectID, platformID)
		return err
	})

	g.Go(func() error {
		var err error
		publishPost, err = s.postService.GetPostToPublish(ctx, postID)
		return err
	})

	g.Go(func() error {
		var err error
		media, err = s.mediaService.GetMediaForPost(ctx, projectID, postID, platformID)
		return err
	})

	g.Go(func() error {
		var err error
		secrets, err = s.repo.GetUserPlatformSecrets(ctx, platformID, defaultUserID)
		return err
	})

	if err := g.Wait(); err != nil {
		return err
	}

	if !isEnabled {
		return ErrSocialPlatformNotEnabledForProject
	}

	if publishPost == nil {
		return post.ErrPostNotFound
	}

	if secrets == "" {
		return ErrUserSecretsNotSet
	}

	publisher, err := s.publisherFactory.Create(platformID, secrets)
	if err != nil {
		return err
	}

	if err := publisher.Publish(ctx, publishPost, media); err != nil {
		fmt.Printf("Failed to publish post to %s: %v\n", platformID, err)
		e := s.postService.UpdatePostStatus(ctx, postID, post.PostStatusFailed)
		if e != nil {
			return fmt.Errorf("failed to update publish post status to failed: %w", e)
		}
		return err
	}

	if err := s.postService.UpdatePostStatus(ctx, postID, post.PostStatusPublished); err != nil {
		return fmt.Errorf("failed to update publish post status to published: %w", err)
	}

	return nil
}

func (s *service) ValidatePostForSocialNetwork(ctx context.Context, projectID, postID, platformID string) error {
	var (
		isEnabled     bool
		publishPost   *post.PublishPost
		media         []*media.Media
		secrets       string
		defaultUserID string
		g             errgroup.Group
	)

	defaultUserID, err := s.repo.GetDefaultUserID(ctx, projectID)
	if err != nil {
		return err
	}
	if defaultUserID == "" {
		return ErrDefaultUserNotSet
	}

	g.Go(func() error {
		var err error
		isEnabled, err = s.repo.IsSocialNetworkEnabledForProject(ctx, projectID, platformID)
		return err
	})

	g.Go(func() error {
		var err error
		publishPost, err = s.postService.GetPostToPublish(ctx, postID)
		return err
	})

	g.Go(func() error {
		var err error
		media, err = s.mediaService.GetMediaForPost(ctx, projectID, postID, platformID)
		return err
	})

	g.Go(func() error {
		var err error
		secrets, err = s.repo.GetUserPlatformSecrets(ctx, platformID, defaultUserID)
		return err
	})

	if err := g.Wait(); err != nil {
		return err
	}

	if !isEnabled {
		return ErrSocialPlatformNotEnabledForProject
	}

	if publishPost == nil {
		return post.ErrPostNotFound
	}

	if secrets == "" {
		return ErrUserSecretsNotSet
	}

	publisher, err := s.publisherFactory.Create(platformID, secrets)
	if err != nil {
		return err
	}

	if err := publisher.ValidatePost(ctx, publishPost, media); err != nil {
		return err
	}

	return nil
}
