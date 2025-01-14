package publisher

import "context"

type Repository interface {
	FindAll(ctx context.Context) ([]Platform, error)
	FindByID(ctx context.Context, id string) (*Platform, error)
	GetPlatformSecrets(ctx context.Context, projectID, socialPlatformID string) (*string, error)
	SetPlatformSecrets(ctx context.Context, projectID, socialPlatformID, secrets string) error
	IsSocialNetworkEnabledForProject(ctx context.Context, projectID, socialPlatformID string) (bool, error)
	GetUserPlatformSecrets(ctx context.Context, platformID, userID string) (string, error)
	SetUserPlatformSecrets(ctx context.Context, platformID, userID, secrets string) error
	GetDefaultUserID(ctx context.Context, platformID string) (string, error)
}
