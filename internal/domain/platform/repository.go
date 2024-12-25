package platform

import "context"

type Repository interface {
	FindAll(ctx context.Context) ([]Platform, error)
	FindByID(ctx context.Context, id string) (*Platform, error)
	AddAPIKey(ctx context.Context, projectID, socialPlatformID, apiKey string) error
	IsSocialNetworkEnabledForProject(ctx context.Context, projectID, socialPlatformID string) (bool, error)
}
