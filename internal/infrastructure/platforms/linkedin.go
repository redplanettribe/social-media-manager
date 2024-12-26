package platforms

import (
	"context"
	"fmt"
	"time"

	"github.com/pedrodcsjostrom/opencm/internal/domain/post"
)

// Linkedin is a struct that represents the Linkedin platform publisher
type Linkedin struct {
	ID string
	apiKey string
}

func (l *Linkedin) Publish(ctx context.Context, post *post.QPost)  error{
    // Publish to Linkedin
    fmt.Println("Publishing to Linkedin")
	time.Sleep(1 * time.Second)
    return nil
}

func NewLinkedin(apiKey string) *Linkedin {
	return &Linkedin{
		ID: "linkedin",
		apiKey: apiKey,
	}
}