package publisher

import (
	"context"
	"fmt"
	"sync"

	"github.com/pedrodcsjostrom/opencm/internal/domain/post"
	"github.com/pedrodcsjostrom/opencm/internal/infrastructure/config"
	"github.com/pedrodcsjostrom/opencm/internal/infrastructure/platforms"
)

// PublisherQueue manages the channels and workers for publishing
type PublisherQueue struct {
	publishCh        chan *post.QPost
	retryCh          chan *post.QPost
	publisherFactory platforms.PublisherFactory
	wg               sync.WaitGroup
	cfg              *config.PublisherConfig
}

// NewPublisherQueue initializes the queue with desired worker counts
func NewPublisherQueue(cfg *config.PublisherConfig) *PublisherQueue {
	return &PublisherQueue{
		publishCh:        make(chan *post.QPost, cfg.PublishBuffer),
		retryCh:          make(chan *post.QPost, cfg.RetryBuffer),
		publisherFactory: platforms.NewPublisherFactory(),
		cfg:              cfg,
		wg:               sync.WaitGroup{},
	}
}

// Start spins up workers for publishing and retry handling
func (pq *PublisherQueue) Start(ctx context.Context) {
	for i := 0; i < pq.cfg.WorkerNum; i++ {
		pq.wg.Add(1)
		go pq.runPublishWorker(ctx)
	}
	for i := 0; i < pq.cfg.RetryNum; i++ {
		pq.wg.Add(1)
		go pq.runRetryWorker(ctx)
	}
}

// Stop signals workers to finish
func (pq *PublisherQueue) Stop() {
	close(pq.publishCh)
	close(pq.retryCh)
	pq.wg.Wait()
}

// Enqueue adds a post to the publishCh
func (pq *PublisherQueue) Enqueue(p *post.QPost) {
	pq.publishCh <- p
}

// runPublishWorker consumes publishCh, on error sends to retryCh
func (pq *PublisherQueue) runPublishWorker(ctx context.Context) {
	defer pq.wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case p, ok := <-pq.publishCh:
			if !ok {
				return
			}
			if err := pq.publishPost(ctx, p); err != nil {
				pq.retryCh <- p
			}
		}
	}
}

// runRetryWorker tries to re-publish failed posts
func (pq *PublisherQueue) runRetryWorker(ctx context.Context) {
	defer pq.wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case p, ok := <-pq.retryCh:
			if !ok {
				return
			}
			if err := pq.publishPost(ctx, p); err != nil {
				fmt.Printf("Post %s failed again: %v\n", p.ID, err)
				// persist post status to db as failed
				fmt.Println("persisting failed post status to db")
				// dequeue
				return
			}
		}
	}
}

// Example publishing logic
func (pq *PublisherQueue) publishPost(ctx context.Context, p *post.QPost) error {
	pub := pq.publisherFactory.Create(p.Platform, p.ApiKey)
	if err := pub.Publish(ctx, p); err != nil {
		return err
	}
	return nil
}
