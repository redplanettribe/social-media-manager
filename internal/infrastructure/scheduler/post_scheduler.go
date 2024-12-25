package scheduler

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/pedrodcsjostrom/opencm/internal/domain/post"
	"github.com/pedrodcsjostrom/opencm/internal/domain/project"
	"github.com/pedrodcsjostrom/opencm/internal/infrastructure/config"
)

type PostScheduler struct {
	postService    post.Service
	projectService project.Service
	cfg            *config.SchedulerConfig
	quit           chan struct{}
}

func NewPostScheduler(
	postSvc post.Service,
	projectSvc project.Service,
	cfg *config.SchedulerConfig,
) *PostScheduler {
	return &PostScheduler{
		postService:    postSvc,
		projectService: projectSvc,
		cfg:            cfg,
		quit:           make(chan struct{}),
	}
}

func (s *PostScheduler) Start(ctx context.Context) {
	ticker := time.NewTicker(s.cfg.Interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				if err := s.processQueue(ctx); err != nil {
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

func (s *PostScheduler) processQueue(_ context.Context) error {
	fmt.Println("Tick")
	// Something like this maybe
	// Get the active projects that have posts in the queue
	// Check if we have any scheduled posts to process
	// Take the first post from the queue it is time to process according project release schedule
	// We'll need a way to track the process and limit the amount of posts we process at a time

	// for _, post := range posts {
	//     // Process each post
	//     if err := s.processPost(ctx, post); err != nil {
	//         log.Printf("Error processing post %s: %v", post.ID, err)
	//         continue
	//     }
	// }
	return nil
}

// func (s *PostScheduler) processPost(ctx context.Context, p *post.Post) error {
// 	// Get all the publishers associated with the post
// 	// for _, publisher := range publishers {
// 	//     // Publish the post

// 	return nil
// }
