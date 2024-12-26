package platform

import (
	"context"

	"golang.org/x/sync/errgroup"
)

type Service interface {
	GetAvailableSocialNetworks(ctx context.Context) ([]Platform, error)
	AddAPIKey(ctx context.Context, projectID, socialNetworkID, apiKey string) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) GetAvailableSocialNetworks(ctx context.Context) ([]Platform, error) {
	return s.repo.FindAll(ctx)
}

func (s *service) AddAPIKey(ctx context.Context, projectID, socialPlatformID, apiKey string) error {
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

	return s.repo.AddAPIKey(ctx, projectID, socialPlatformID, apiKey)
}
