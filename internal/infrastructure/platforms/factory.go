package platforms

import (
	"errors"

	"github.com/pedrodcsjostrom/opencm/internal/domain/publisher"
	"github.com/pedrodcsjostrom/opencm/internal/infrastructure/encrypting"
	"github.com/pedrodcsjostrom/opencm/internal/infrastructure/platforms/linkedin"
)

// NewPublisherFactory creates a new PublisherFactory
func NewPublisherFactory(e encrypting.Encrypter) *publisherFactory {
	return &publisherFactory{
		encrypter: e,
	}
}

type publisherFactory struct {
	encrypter encrypting.Encrypter
}

func (pf *publisherFactory) Create(platform string, platformSecrets, userSecrets string) (publisher.Publisher, error) {
	var p publisher.Publisher
	e := pf.encrypter
	switch platform {
	case "linkedin":
		p = linkedin.NewLinkedin(platformSecrets, e)
	default:
		return nil, errors.New("unknown platform")
	}

	err := p.ValidatePlatformSecrets(platformSecrets)
	if err != nil {
		return nil, err
	}

	err = p.ValidateUserSecrets(userSecrets)
	if err != nil {
		return nil, err
	}

	return p, nil
}
