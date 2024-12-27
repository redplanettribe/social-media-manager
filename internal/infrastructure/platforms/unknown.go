package platforms

import (
	"context"
	"errors"

	"github.com/pedrodcsjostrom/opencm/internal/domain/post"
)

type UnknownPublisher struct {
}

func NewUnknownPublisher() *UnknownPublisher {
	return &UnknownPublisher{}
}

func (p *UnknownPublisher) Publish(ctx context.Context, post *post.QPost) error {
	return errors.New("unknown platform")
}