package linkedin

import (
	"context"
	"errors"

	"github.com/pedrodcsjostrom/opencm/internal/domain/media"
	"github.com/pedrodcsjostrom/opencm/internal/domain/post"
)

type Poster interface {
	Post(ctx context.Context, post *post.PublishPost, media []*media.Media) error
}

type PosterFactory interface {
	NewPoster(p *post.PublishPost, userSecrets UserSecrets, platformSecrets PlatformSecrets) (Poster, error)
}

type posterFactory struct {
}

func NewPosterFactory() PosterFactory {
	return &posterFactory{}
}

func (pf *posterFactory) NewPoster(p *post.PublishPost, userSecrets UserSecrets, platformSecrets PlatformSecrets) (Poster, error) {
	if p.Type == post.PostTypeUndefined {
		return nil, errors.New("post type is undefined")
	}
	switch p.Type {
	case post.PostTypeText:
		return NewTextPoster(userSecrets, platformSecrets), nil
	default:
		return nil, errors.New("invalid post type")
	}
}

type TextPoster struct {
	uSecrets UserSecrets
	pSecrets PlatformSecrets
}

func NewTextPoster(userSecrets UserSecrets, platformSecrets PlatformSecrets ) *TextPoster {
	return &TextPoster{
		uSecrets: userSecrets,
		pSecrets: platformSecrets,
	}
}

func (tp *TextPoster) Post(ctx context.Context, post *post.PublishPost, media []*media.Media) error {
	return nil
}
