package platform

import "context"

type Repository interface {
	FindAll(ctx context.Context) ([]Platform, error)
	FindByID(ctx context.Context, id string) (*Platform, error)
	GetSecrets(ctx context.Context, projectID, socialPlatformID string) (*string, error)
	SetSecrets(ctx context.Context, projectID, socialPlatformID, secrets string) error
	IsSocialNetworkEnabledForProject(ctx context.Context, projectID, socialPlatformID string) (bool, error)
}
