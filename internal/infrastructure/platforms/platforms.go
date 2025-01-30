package platforms

import (
	"errors"

	"github.com/pedrodcsjostrom/opencm/internal/domain/publisher"
	"github.com/pedrodcsjostrom/opencm/internal/infrastructure/encrypting"
	"github.com/pedrodcsjostrom/opencm/internal/infrastructure/platforms/linkedin"
)

var (
	ErrUnknownPlatform = errors.New("unknown platform")
	ErrNotImplemented  = errors.New("not implemented")
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

func (pf *publisherFactory) Create(platform string, secrets string) (publisher.Publisher, error) {
	var p publisher.Publisher
	e := pf.encrypter
	switch platform {
	case "linkedin":
		p = linkedin.NewLinkedin(secrets, e)
	default:
		return nil, errors.New("unknown platform")
	}
	return p, nil
}
