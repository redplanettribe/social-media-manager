package scheduler

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/pedrodcsjostrom/opencm/internal/domain/post"
	"github.com/pedrodcsjostrom/opencm/internal/domain/project"
	"github.com/pedrodcsjostrom/opencm/internal/infrastructure/config"
	pq "github.com/pedrodcsjostrom/opencm/internal/infrastructure/publisher_queue"
	"golang.org/x/sync/errgroup"
)

type PostScheduler struct {
	postService    post.Service
	projectService project.Service
	cfg            *config.SchedulerConfig
	publisherQueue pq.PublisherQueue
	quit           chan struct{}
}

func NewPostScheduler(
	postSvc post.Service,
	projectSvc project.Service,
	publisherQueue pq.PublisherQueue,
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
					log.Printf("Error S: %v", err)
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

// scanAndEnqueue orchestrates concurrent, chunked queries for posts.
// It combines posts into a single channel, deduplicates them, then enqueues each.
func (s *PostScheduler) scanAndEnqueue(ctx context.Context) error {
	// Channel to collect QPost from multiple scanners
	qPosts := make(chan *post.QPost, s.cfg.ChannelBuffer)

	// Use errgroup for concurrency & combined error handling
	g, gCtx := errgroup.WithContext(ctx)

	// Scanner #1: scheduled posts ready for publication
	g.Go(func() error {
		return s.scanScheduledPosts(gCtx, qPosts)
	})

	// Scanner #2: project post queues
	g.Go(func() error {
		return s.scanProjectQueues(gCtx, qPosts)
	})

	// Once both scanners are done, close channel
	go func() {
		_ = g.Wait() // we ignore the error here, handled when g.Wait() is called below
		close(qPosts)
	}()

	// Deduplicate and enqueue
	processed := make(map[string]bool)
	for q := range qPosts {
		// Deduplicate based on (PostID + Platform)
		sig := fmt.Sprintf("%s|%s", q.ID, q.Platform)
		if processed[sig] {
			continue
		}
		processed[sig] = true

		s.publisherQueue.Enqueue(ctx, q)
	}

	// Wait for the scanners to conclude
	if err := g.Wait(); err != nil {
		return err
	}
	return nil
}

// scanScheduledPosts queries posts that are directly scheduled (e.g., with a scheduled_at).
// It pages through results in chunks to avoid huge queries all at once.
func (s *PostScheduler) scanScheduledPosts(ctx context.Context, out chan<- *post.QPost) error {
	const chunkSize = 100
	offset := 0

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Retrieve a chunk of scheduled posts
		chunk, err := s.postService.FindScheduledReadyPosts(ctx, offset, chunkSize)
		fmt.Printf("Found %d scheduled posts\n", len(chunk))
		for _, p := range chunk {
			fmt.Printf("Post %s: %s\n", p.ID, p.Title)
		}
		if err != nil {
			return err
		}
		if len(chunk) == 0 {
			break
		}

		// Send each post to the out channel
		for _, p := range chunk {
			out <- p
		}

		offset += chunkSize
	}
	return nil
}

// scanProjectQueues queries for posts in each project's custom queue that are due to be published.
// Also pages through results in chunks to avoid big loads.
func (s *PostScheduler) scanProjectQueues(ctx context.Context, out chan<- *post.QPost) error {
	// chunkSize for projects
	const chunkSize = 20
	offset := 0

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Retrieve a chunk of active projects
		projs, err := s.projectService.FindActiveProjectsChunk(ctx, offset, chunkSize)
		if err != nil {
			return err
		}
		if len(projs) == 0 {
			break
		}

		// We'll gather posts from each chunk of projects concurrently
		g, gCtx := errgroup.WithContext(ctx)

		for _, proj := range projs {
			projectID := proj.ID
			g.Go(func() error {
				// Check if it's time to publish for this project according to its configured schedule
				ok, err := s.projectService.IsProjectTimeToPublish(gCtx, projectID)
				if err != nil {
					return err
				}
				if !ok {
					return nil // not time to publish
				}

				// Each post can have multiple platforms to publish
				qps, err := s.postService.DequeuePostsToPublish(gCtx, projectID)
				if err != nil {
					return err
				}
				if qps == nil {
					return nil // no post to enqueue
				}

				// Send each post to the out channel to enqueue
				for _, qp := range qps {
					out <- qp
				}
				return nil
			})
		}

		// Wait for all projects in this chunk
		if err := g.Wait(); err != nil {
			return err
		}

		offset += chunkSize
	}

	return nil
}
