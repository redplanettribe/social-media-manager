package platforms

import (
	"context"

	"github.com/pedrodcsjostrom/opencm/internal/domain/post"
)

//go:generate mockery --name=Publisher --case=underscore --inpackage
type Publisher interface {
	Publish(ctx context.Context, post *post.QPost) error
}


// PublisherFactory is a factory for creating publishers
//go:generate mockery --name=PublisherFactory --case=underscore --inpackage
type PublisherFactory interface {
    Create(platform string, apiKey string) Publisher
}

// NewPublisherFactory creates a new PublisherFactory
func NewPublisherFactory() PublisherFactory {
    return &publisherFactory{}
}

type publisherFactory struct {}

func (f *publisherFactory) Create(platform string, apiKey string) Publisher {
    switch platform {
    case "linkedin":
        return NewLinkedin(apiKey)
    case "instagram":
        return NewInstagram(apiKey)
    case "x":
        return NewX(apiKey)
    default:
        return NewUnknownPublisher()
    }
}
