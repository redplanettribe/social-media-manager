package publisher

import (
	"context"
	"errors"

	"github.com/pedrodcsjostrom/opencm/internal/domain/media"
	"github.com/pedrodcsjostrom/opencm/internal/domain/post"
	"github.com/pedrodcsjostrom/opencm/internal/infrastructure/encrypting"
	"github.com/pedrodcsjostrom/opencm/internal/infrastructure/platforms"
)

//go:generate mockery --name=Publisher --case=underscore --inpackage
type Publisher interface {
    // Publish a post with media to the platform. Media could be nil
	Publish(ctx context.Context, post *post.PublishPost, media []*media.Media) error
    // Check if the secrets are valid
    ValidateSecrets(Secrets string) error
    // Add a new secret and return the encrypted secrets string. If the key already exists, it will be updated. If the key is not valid for this platform, an error will be returned
    AddSecret(key, secret string) (string, error)
}


// PublisherFactory is a factory for creating publishers
//go:generate mockery --name=PublisherFactory --case=underscore --inpackage
type PublisherFactory interface {
    Create(platform string, secrets string) (Publisher, error)
}

// NewPublisherFactory creates a new PublisherFactory
func NewPublisherFactory( e encrypting.Encrypter) PublisherFactory {
    return &publisherFactory{
        encrypter: e,
    }
}

type publisherFactory struct {
    encrypter encrypting.Encrypter
}

func (pf *publisherFactory) Create(platform string, secrets string) (Publisher, error) {
    var p Publisher
    e := pf.encrypter
    switch platform {
    case "linkedin":
        p = platforms.NewLinkedin(secrets, e)
    default:
        return nil, errors.New("unknown platform")
    }

    err:= p.ValidateSecrets(secrets)
    if err != nil {
        return nil, err
    }

    return p, nil
}
