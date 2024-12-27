package platforms

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/pedrodcsjostrom/opencm/internal/domain/post"
)

type X struct {
    ID string
	apiKey string
}

func (t *X) Publish(ctx context.Context, post *post.QPost) error {
    // Publish to Twitter
    fmt.Println("Publishing to X")
	time.Sleep(1 * time.Second)
	return errors.New("X platform not implemented")
}

func NewX(apiKey string) *X {
	return &X{
		ID: "x",
		apiKey: apiKey,
	}
}
