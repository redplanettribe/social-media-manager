package publisher

import (
	"context"

	"github.com/pedrodcsjostrom/opencm/internal/domain/media"
	post "github.com/pedrodcsjostrom/opencm/internal/domain/post"
	"github.com/pedrodcsjostrom/opencm/internal/infrastructure/encrypting"
	"golang.org/x/sync/errgroup"
)

type Service interface {
	GetAvailableSocialNetworks(ctx context.Context) ([]Platform, error)
	AddSecret(ctx context.Context, projectID, socialNetworkID, key, secret string) error
	PublishPost(ctx context.Context, postID string) error
	PublishPostToSocialNetwork(ctx context.Context, projectID, postID, socialNetworkID string) error
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

func (s *service) AddSecret(ctx context.Context, projectID, socialPlatformID, key, secret string) error {
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
		encryptedSecrets, err = s.repo.GetSecrets(ctx, projectID, socialPlatformID)
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

	publisher, err := s.publisherFactory.Create(socialPlatformID, *encryptedSecrets)
	if err != nil {
		return err
	}

	newSecrets, err := publisher.AddSecret(key, secret)
	if err != nil {
		return err
	}

	return s.repo.SetSecrets(ctx, projectID, socialPlatformID, newSecrets)
}

func (s *service) PublishPost(ctx context.Context, postID string) error {
	return nil
}

func (s *service) PublishPostToSocialNetwork(ctx context.Context, projectID, postID, socialNetworkID string) error {
	var (
		isEnabled bool
		p         *post.PublishPost
		g         errgroup.Group
	)

	g.Go(func() error {
		var err error
		isEnabled, err = s.repo.IsSocialNetworkEnabledForProject(ctx, projectID, socialNetworkID)
		return err
	})

	g.Go(func() error {
		var err error
		p, err = s.postService.GetPostToPublish(ctx, postID)
		return err
	})

	if err := g.Wait(); err != nil {
		return err
	}

	if !isEnabled {
		return ErrSocialPlatformNotEnabledForProject
	}

	if p == nil {
		return post.ErrPostNotFound
	}

	publisher, err := s.publisherFactory.Create(socialNetworkID, p.Secrets)
	if err != nil {
		return err
	}

	if err := publisher.Publish(ctx, p); err != nil {
		return err
	}

	return nil
}
