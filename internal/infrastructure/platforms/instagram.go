package platforms

import (
	"context"
	"fmt"
	"time"

	"github.com/pedrodcsjostrom/opencm/internal/domain/post"
)

type Instagram struct {
	ID string
	apiKey string
}

func (f *Instagram) Publish(ctx context.Context, post *post.QPost) error {
	// Publish to Facebook
	fmt.Println("Publishing to Instagram")
	time.Sleep(1 * time.Second)
	return nil
}


func NewInstagram(apikey string) *Instagram {
	return &Instagram{
		ID: "instagram",
		apiKey: apikey,
	}
}
