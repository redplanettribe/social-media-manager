package x

import (
	"context"
	"errors"

	"github.com/pedrodcsjostrom/opencm/internal/domain/media"
	"github.com/pedrodcsjostrom/opencm/internal/domain/post"
)

type XPoster interface {
	Post(ctx context.Context, post *post.PublishPost, media []*media.Media) error
	Validate(ctx context.Context, post *post.PublishPost, media []*media.Media) error
}

type XPosterFactory interface {
	NewPoster(p *post.PublishPost, userSecrets Secrets) (XPoster, error)
}

type posterFactory struct{}

func NewXPosterFactory() XPosterFactory {
	return &posterFactory{}
}

func (pf *posterFactory) NewPoster(p *post.PublishPost, secrets Secrets) (XPoster, error) {
	switch p.Type {
	case post.PostTypeText:
		return NewTextPoster(secrets), nil
	case post.PostTypeImage, post.PostTypeMultiImage:
		return NewMediaPoster(secrets), nil
	default:
		return nil, errors.New("invalid post type")
	}
}
