package publisher

import (
	"context"

	"github.com/pedrodcsjostrom/opencm/internal/domain/media"
	"github.com/pedrodcsjostrom/opencm/internal/domain/post"
	"github.com/pedrodcsjostrom/opencm/internal/infrastructure/platforms"
)

type Service interface {
	Publish(ctx context.Context, postID string) error
}

type service struct {
	postService post.Service
	mediaService media.Service
	publisherFactory platforms.PublisherFactory
}