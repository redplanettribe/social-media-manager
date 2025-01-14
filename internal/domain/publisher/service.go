package publisher

import (
	"context"

	"github.com/pedrodcsjostrom/opencm/internal/domain/media"
	post "github.com/pedrodcsjostrom/opencm/internal/domain/post"
	"github.com/pedrodcsjostrom/opencm/internal/infrastructure/encrypting"
	"github.com/pedrodcsjostrom/opencm/internal/interfaces/api/http/middlewares"
	"golang.org/x/sync/errgroup"
)

//go:generate mockery --name=Service --case=underscore --inpackage
type Service interface {
	GetAvailableSocialNetworks(ctx context.Context) ([]Platform, error)
	AddPlatformSecret(ctx context.Context, projectID, platformID, key, secret string) error
	PublishPostToAssignedSocialNetworks(ctx context.Context, projecID, postID string) error
	PublishPostToSocialNetwork(ctx context.Context, projectID, postID, platformID string) error
	AddUserPlatformSecret(ctx context.Context, projectID, platformID, key, secret string) error
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

func (s *service) GetAvailableSocialNetworks(ctx context.Context) ([]Platform, error) {
	return s.repo.FindAll(ctx)
}

func (s *service) AddPlatformSecret(ctx context.Context, projectID, socialPlatformID, key, secret string) error {
	var (
		sp               *Platform
		encryptedSecrets *string
		isEnabled        bool
		g                errgroup.Group
	)

	g.Go(func() error {
		var err error
		sp, err = s.repo.FindByID(ctx, socialPlatformID)
		return err
	})

	g.Go(func() error {
		var err error
		isEnabled, err = s.repo.IsSocialNetworkEnabledForProject(ctx, projectID, socialPlatformID)
		return err
	})

	g.Go(func() error {
		var err error
		encryptedSecrets, err = s.repo.GetPlatformSecrets(ctx, projectID, socialPlatformID)
		if encryptedSecrets == nil {
			encryptedSecrets = new(string)
		}
		return err
	})
	if err := g.Wait(); err != nil {
		return err
	}

	if sp == nil {
		return ErrSocialPlatformNotFound
	}

	if !isEnabled {
		return ErrSocialPlatformNotEnabledForProject
	}

	publisher, err := s.publisherFactory.Create(socialPlatformID, *encryptedSecrets, "")
	if err != nil {
		return err
	}

	newSecrets, err := publisher.AddPlatformSecret(key, secret)
	if err != nil {
		return err
	}

	return s.repo.SetPlatformSecrets(ctx, projectID, socialPlatformID, newSecrets)
}

func (s *service) AddUserPlatformSecret(ctx context.Context, projectID, platformID, key, secret string) error {
	userID := ctx.Value(middlewares.UserIDKey).(string)

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

	publisher, err := s.publisherFactory.Create(platformID, "", userSecrets)
	if err != nil {
		return err
	}

	newSecrets, err := publisher.AddUserSecret(key, secret)
	if err != nil {
		return err
	}

	return s.repo.SetUserPlatformSecrets(ctx, platformID, userID, newSecrets)
}

func (s *service) PublishPostToAssignedSocialNetworks(ctx context.Context, projecID, postID string) error {

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
			return s.PublishPostToSocialNetwork(gCtx, projecID, postID, pid)
		})
	}

	return g.Wait()
}

func (s *service) PublishPostToSocialNetwork(ctx context.Context, projectID, postID, platformID string) error {
	var (
		isEnabled     bool
		publishPost   *post.PublishPost
		media         []*media.Media
		userSecrets   string
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
		userSecrets, err = s.repo.GetUserPlatformSecrets(ctx, platformID, defaultUserID)
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

	if userSecrets == "" {
		return ErrUserSecretsNotSet
	}

	publisher, err := s.publisherFactory.Create(platformID, publishPost.Secrets, userSecrets)
	if err != nil {
		return err
	}

	if err := publisher.Publish(ctx, publishPost, media); err != nil {
		return err
	}

	return nil
}
