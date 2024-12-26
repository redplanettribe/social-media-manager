package scheduler

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/pedrodcsjostrom/opencm/internal/domain/post"
	"github.com/pedrodcsjostrom/opencm/internal/domain/project"
	"github.com/pedrodcsjostrom/opencm/internal/infrastructure/config"
	"github.com/pedrodcsjostrom/opencm/internal/infrastructure/publisher"
)

type PostScheduler struct {
	postService    post.Service
	projectService project.Service
	cfg            *config.SchedulerConfig
	publisherQueue *publisher.PublisherQueue
	quit           chan struct{}
}

func NewPostScheduler(
	postSvc post.Service,
	projectSvc project.Service,
	publisherQueue *publisher.PublisherQueue,
	cfg *config.SchedulerConfig,
) *PostScheduler {
	return &PostScheduler{
		postService:    postSvc,
		projectService: projectSvc,
		cfg:            cfg,
		publisherQueue: publisherQueue,
		quit:           make(chan struct{}),
	}
}

func (s *PostScheduler) Start(ctx context.Context) {
	ticker := time.NewTicker(s.cfg.Interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				if err := s.scanAndEnqueue(ctx); err != nil {
					log.Printf("Error processing queue: %v", err)
				}
			case <-s.quit:
				ticker.Stop()
				return
			case <-ctx.Done():
				ticker.Stop()
				return
			}
		}
	}()
}

func (s *PostScheduler) Stop() {
	close(s.quit)
}

func (s *PostScheduler) scanAndEnqueue(_ context.Context) error {
	fmt.Println("Tick")
	// Something like this maybe
	// Get the active projects that have posts in the queue
	// Check if we have any scheduled posts to process
	// Take the first post from the queue it is time to process according project release schedule
	// We'll need a way to track the process and limit the amount of posts we process at a time
	posts:= []*post.QPost{
		{
			Post: &post.Post{ 
				ID: "1",
				Title: "Post 1",
				TextContent: "Post 1 content",
			},
			Platform: "x",
			ApiKey: "x-api-key",
		},
		{
			Post: &post.Post{ 
				ID: "1",
				Title: "Post 1",
				TextContent: "Post 1 content",
			},
			Platform: "instagram",
			ApiKey: "instagram-api-key",
		},
		{
			Post: &post.Post{ 
				ID: "1",
				Title: "Post 1",
				TextContent: "Post 1 content",
			},
			Platform: "linkedin",
			ApiKey: "linkedin-api-key",
		},
	}
	for _, post := range posts {
	    // Process each post
		s.publisherQueue.Enqueue(post)
	}
	return nil
}