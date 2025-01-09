package platform

import (
	"context"

	"github.com/pedrodcsjostrom/opencm/internal/infrastructure/encrypting"
	"github.com/pedrodcsjostrom/opencm/internal/infrastructure/platforms"
	"golang.org/x/sync/errgroup"
)

type Service interface {
	GetAvailableSocialNetworks(ctx context.Context) ([]Platform, error)
	AddSecret(ctx context.Context, projectID, socialNetworkID, key, secret string) error
}

type service struct {
	repo             Repository
	publisherFactory platforms.PublisherFactory
	encrypter        encrypting.Encrypter
}

func NewService(r Repository, e encrypting.Encrypter, pf platforms.PublisherFactory) Service {
	return &service{
		repo:             r,
		publisherFactory: pf,
		encrypter:        e,
	}
}

func (s *service) GetAvailableSocialNetworks(ctx context.Context) ([]Platform, error) {
	return s.repo.FindAll(ctx)
}

func (s *service) AddSecret(ctx context.Context, projectID, socialPlatformID, key, secret string) error {
	var (
		sp        *Platform
		isEnabled bool
		g         errgroup.Group
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

	if err := g.Wait(); err != nil {
		return err
	}

	if sp == nil {
		return ErrSocialPlatformNotFound
	}

	if !isEnabled {
		return ErrSocialPlatformNotEnabledForProject
	}

	encryptedSecrets, err := s.repo.GetSecrets(ctx, projectID, socialPlatformID)
	if err != nil {
		return err
	}
	if encryptedSecrets == nil {
		encryptedSecrets = new(string)
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
