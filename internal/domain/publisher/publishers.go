package publisher

import (
	"context"

	"github.com/pedrodcsjostrom/opencm/internal/domain/media"
	"github.com/pedrodcsjostrom/opencm/internal/domain/post"
)

//go:generate mockery --name=Publisher --case=underscore --inpackage
type Publisher interface {
    // Publish a post with media to the platform. Media could be nil
	Publish(ctx context.Context, post *post.PublishPost, media []*media.Media) error
    // Check if the platform secrets are valid
    ValidatePlatformSecrets(secrets string) error
    // Check if the user secrets are valid
    ValidateUserSecrets(secrets string) error
    // Add a new platform secret and return the encrypted secrets string. If the key already exists, it will be updated. If the key is not valid for this platform, an error will be returned
    AddPlatformSecret(key, secret string) (string, error)
    // Add a new user secret and return the encrypted secrets string. If the key already exists, it will be updated. If the key is not valid for this platform, an error will be returned
    AddUserSecret(key, secret string) (string, error)
}


// PublisherFactory is a factory for creating publishers
//go:generate mockery --name=PublisherFactory --case=underscore --inpackage
type PublisherFactory interface {
    Create(platform string, platformSecrets, userSecrets string) (Publisher, error)
}

