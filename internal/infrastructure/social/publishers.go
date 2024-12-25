package social

import "context"

type Publisher interface {
    Publish(ctx context.Context, content string, mediaURLs []string) error
}

type TwitterPublisher struct {
    // Twitter API client configuration
}

type LinkedinPublisher struct {
    // Facebook API client configuration
}

// Implement other social media publishers here