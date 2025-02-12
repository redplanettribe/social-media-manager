package publisher

import (
	"context"
	"time"

	"github.com/pedrodcsjostrom/opencm/internal/domain/media"
	"github.com/pedrodcsjostrom/opencm/internal/domain/post"
)

//go:generate mockery --name=Publisher --case=underscore --inpackage
type Publisher interface {
	// Authenticate the user and return the encrypted secrets string, plus the token expiration time. If the code is invalid, an error will be returned
	Authenticate(ctx context.Context, params any) (string, time.Time, error)
	// Publish a post with media to the platform. Media could be nil
	Publish(ctx context.Context, post *post.PublishPost, media []*media.Media) error
	// Validate Post
	ValidatePost(ctx context.Context, post *post.PublishPost, media []*media.Media) error
}

//go:generate mockery --name=PublisherFactory --case=underscore --inpackage
type PublisherFactory interface {
	Create(platform string, secrets string) (Publisher, error)
}

type PublishPostInfo struct {
	Post  *post.PublishPost
	Media []*media.DownloadMetaData
}
